package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/print-foundation/print/internal/logging"
)

func assembleISO(work, isoPath string, spec Spec, log logging.Logger) error {
	client := filepath.Join(work, "print-client")

	bootDir := filepath.Join(work, "boot", "grub")
	if err := os.MkdirAll(bootDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(bootDir, "grub.cfg"), []byte(buildGrubCfg(spec, client)), 0o644); err != nil {
		return err
	}

	tools := spec.Tools
	if tools.GrubMkRescue == "" {
		tools.GrubMkRescue = lookPath("grub-mkrescue")
	}
	if tools.Xorriso == "" {
		tools.Xorriso = lookPath("xorriso")
	}

	switch {
	case tools.GrubMkRescue != "":
		cmd := exec.Command(tools.GrubMkRescue, "-o", isoPath, work)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("grub-mkrescue failed: %w\n%s", err, out)
		}
		log.Info("iso assembled with grub-mkrescue", "path", isoPath)
		return nil
	case tools.Xorriso != "":
		cmd := exec.Command(tools.Xorriso, "-as", "mkisofs",
			"-iso-level", "3", "-full-iso9660-filenames",
			"-b", "boot/grub/i386-pc/eltorito.img", "-no-emul-boot",
			"-boot-load-size", "4", "-boot-info-table",
			"-o", isoPath, work)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("xorriso failed: %w\n%s", err, out)
		}
		log.Info("iso assembled with xorriso", "path", isoPath)
		return nil
	default:
		return fmt.Errorf("%w: need grub-mkrescue or xorriso on PATH to build the ISO", errNoISOTool)
	}
}

func buildGrubCfg(spec Spec, client string) string {
	return fmt.Sprintf(`set timeout=5
set default=0

menuentry "Pr!nt %s installer (%s)" {
    linux /kernel %s
    initrd /initrd
}
`, spec.Distro, spec.Arch, kernelCmdline(client))
}

func kernelCmdline(client string) string {
	return "print.client=" + client + " print.config=/print-client.json quiet"
}

func lookPath(name string) string {
	p, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return p
}

var errNoISOTool = fmt.Errorf("no ISO assembler")
