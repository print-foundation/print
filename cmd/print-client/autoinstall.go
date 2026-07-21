package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/print-foundation/print/internal/builder"
	"github.com/print-foundation/print/internal/logging"
	"github.com/print-foundation/print/internal/network"
)

func runAutoinstall(log logging.Logger) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	net := network.New(log)
	if cfg.WiFi.SSID != "" {
		if err := net.ConnectWiFi(context.Background(), "", cfg.WiFi.SSID, cfg.WiFi.Passphrase); err != nil {
			return fmt.Errorf("wifi connect: %w", err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if !network.NewConnectivity().Online(ctx) {
		return fmt.Errorf("no internet connectivity; cannot install from cloud")
	}

	if _, err := os.Stat(seedPath); err != nil {
		seed, gerr := generateSeed(cfg)
		if gerr != nil {
			return fmt.Errorf("generate seed: %w", gerr)
		}
		if werr := os.WriteFile(seedPath, []byte(seed), 0o644); werr != nil {
			return fmt.Errorf("write seed: %w", werr)
		}
	}

	log.Info("starting unattended install", "distro", cfg.Distro, "seed", seedPath)
	return launchInstaller(cfg)
}

func launchInstaller(cfg builder.ClientConfig) error {
	switch cfg.Distro {
	case "debian", "ubuntu":
		return exec.Command("debian-installer", "--automatic", "--preseedfile", seedPath).Run()
	case "fedora":
		if p, err := exec.LookPath("anaconda"); err == nil {
			return exec.Command(p, "--kickstart", seedPath).Run()
		}
		return fmt.Errorf("anaconda not found in this environment")
	case "arch":
		if p, err := exec.LookPath("archinstall"); err == nil {
			return exec.Command(p, "--config", seedPath).Run()
		}
		return fmt.Errorf("archinstall not found in this environment")
	case "alpine", "freebsd", "openbsd":
		fmt.Fprintf(os.Stdout, "unattended installer for %s is interactive; seed at %s (mirror %s)\n",
			cfg.Distro, seedPath, cfg.Mirror)
		return nil
	default:
		return fmt.Errorf("no installer launcher for %s", cfg.Distro)
	}
}
