package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/print-foundation/print/internal/builder"
	"github.com/print-foundation/print/internal/disk"
	"github.com/print-foundation/print/internal/install"
	"github.com/print-foundation/print/internal/logging"
	"github.com/print-foundation/print/internal/mirrors"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "print:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		flagVersion = flag.Bool("version", false, "print version and exit")
		flagList    = flag.Bool("list-distros", false, "list buildable distros and exit")
		flagTUI     = flag.Bool("tui", false, "launch the interactive TUI wizard")
		flagDistro  = flag.String("distro", "", "distro to build")
		flagCountry = flag.String("country", "", "ISO-3166 country code (e.g. US, DE)")
		flagArch    = flag.String("arch", runtime.GOARCH, "target arch: amd64 or arm64")
		flagOut     = flag.String("out", "", "output .iso path")
		flagWiFi    = flag.String("wifi-ssid", "", "preconfigure WiFi SSID")
		flagPass    = flag.String("wifi-pass", "", "preconfigure WiFi passphrase")
		flagMirror  = flag.String("mirror", "", "explicit mirror URL")
		flagWrite   = flag.Bool("write", false, "flash ISO to device after building")
		flagISO     = flag.String("iso", "", "with -write: path to ISO to flash")
		flagDevice  = flag.String("device", "", "with -write: target device path")
		flagYes     = flag.Bool("yes", false, "with -write: acknowledge destructive write")
	)
	flag.Parse()

	if *flagVersion {
		fmt.Println("Pr!nt", version)
		return nil
	}
	if *flagList {
		for _, d := range mirrors.ListDistros() {
			fmt.Println(string(d))
		}
		fmt.Println("\n(not buildable: windows, darwin/macos need licensed media)")
		return nil
	}

	log := logging.New(logging.Options{Level: "info"})

	if *flagTUI || (*flagDistro == "" && *flagISO == "" && !*flagWrite) {
		return runTUI(log)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if *flagWrite {
		iso := *flagISO
		if iso == "" && *flagOut != "" {
			iso = *flagOut
		}
		if iso == "" || *flagDevice == "" {
			return fmt.Errorf("-write needs -iso (or -out) and -device")
		}
		eng := install.NewEngine(nil, nil, install.WithLogger(log), install.WithDeviceWriter(install.FileDeviceWriter{}))
		err := eng.FlashLocal(ctx, iso, *flagDevice, disk.Confirmation{DevicePath: *flagDevice, Acknowledged: *flagYes}, nil)
		if err != nil {
			return fmt.Errorf("flash: %w", err)
		}
		fmt.Println("flashed", iso, "to", *flagDevice)
		return nil
	}

	if *flagDistro == "" {
		return fmt.Errorf("specify -distro (or run with -tui)")
	}
	distro := mirrors.Distro(*flagDistro)
	if !mirrors.Supported(distro) {
		return fmt.Errorf("%s cannot be built as a cloud ISO (windows/macos need licensed media)", distro)
	}
	arch := builder.Arch(*flagArch)
	if arch != builder.ArchAmd64 && arch != builder.ArchArm64 {
		return fmt.Errorf("unsupported arch %q (use amd64 or arm64)", *flagArch)
	}
	if _, ok := builder.AssetFor(distro, arch); !ok {
		return fmt.Errorf("%s/%s has no netboot asset defined", distro, arch)
	}

	resolver := mirrors.NewResolver()
	var mirror mirrors.Mirror
	if *flagMirror != "" {
		mirror = mirrors.Mirror{BaseURL: *flagMirror, Host: hostOf(*flagMirror)}
	} else {
		country := *flagCountry
		if country == "" {
			country = "US"
		}
		ms, err := resolver.Resolve(ctx, distro, country)
		if err != nil {
			return fmt.Errorf("resolve mirror: %w", err)
		}
		mirror = ms[0]
		log.Info("selected mirror", "host", mirror.Host, "country", mirror.Country)
	}

	clientBin, err := buildClient(log, arch)
	if err != nil {
		return fmt.Errorf("build client: %w", err)
	}

	out := *flagOut
	if out == "" {
		out = filepath.Join(".", "print-"+string(distro)+".iso")
	}

	spec := builder.Spec{
		Distro: distro,
		Arch:   arch,
		Mirror: mirror,
		Output: out,
		WiFi:   builder.WiFiConfig{SSID: *flagWiFi, Passphrase: *flagPass},
	}
	log.Info("building ISO", "distro", distro, "arch", arch, "out", out)
	result, err := builder.New(builder.WithLogger(log)).Build(ctx, spec, clientBin)
	if err != nil {
		return fmt.Errorf("build iso: %w", err)
	}
	fmt.Printf("built %s (verified=%v)\n", result.ISO, result.Verified)

	if *flagWrite && *flagDevice != "" {
		eng := install.NewEngine(nil, nil, install.WithLogger(log), install.WithDeviceWriter(install.FileDeviceWriter{}))
		if err := eng.FlashLocal(ctx, result.ISO, *flagDevice, disk.Confirmation{DevicePath: *flagDevice, Acknowledged: *flagYes}, nil); err != nil {
			return fmt.Errorf("flash: %w", err)
		}
		fmt.Println("flashed", result.ISO, "to", *flagDevice)
	}
	return nil
}

func buildClient(log logging.Logger, arch builder.Arch) ([]byte, error) {
	tmp, err := os.CreateTemp("", "print-client-*")
	if err != nil {
		return nil, err
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	cmd := exec.Command("go", "build", "-o", tmp.Name(), "./cmd/print-client")
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+string(arch))
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("go build client: %w\n%s", err, out)
	}
	log.Info("compiled in-ISO client", "arch", arch)
	return os.ReadFile(tmp.Name())
}

func hostOf(u string) string {
	u = trimPrefix(u, "https://")
	u = trimPrefix(u, "http://")
	if i := indexByte(u, '/'); i >= 0 {
		u = u[:i]
	}
	return u
}

func trimPrefix(s, p string) string {
	if len(s) >= len(p) && s[:len(p)] == p {
		return s[len(p):]
	}
	return s
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
