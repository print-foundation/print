package network

import (
	"bytes"
	"context"
	"net"
	"os/exec"
	"strings"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type baseManager struct {
	log logging.Logger
}

func (b baseManager) Interfaces() ([]domain.NetworkInterface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var out []domain.NetworkInterface
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		ni := domain.NetworkInterface{
			Name:         ifc.Name,
			HardwareAddr: ifc.HardwareAddr.String(),
			Up:           ifc.Flags&net.FlagUp != 0,
			Wireless:     nameLooksWireless(ifc.Name),
		}
		if addrs, err := ifc.Addrs(); err == nil {
			for _, a := range addrs {
				ni.Addresses = append(ni.Addresses, a.String())
			}
		}
		out = append(out, ni)
	}
	return out, nil
}

func nameLooksWireless(name string) bool {
	lower := strings.ToLower(name)
	for _, hint := range []string{"wlan", "wlp", "wi-fi", "wifi", "wireless", "airport"} {
		if strings.Contains(lower, hint) {
			return true
		}
	}
	return false
}

func runCommand(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", &cmdError{cmd: name, msg: strings.TrimSpace(stderr.String()), err: err}
	}
	return stdout.String(), nil
}

type cmdError struct {
	cmd string
	msg string
	err error
}

func (e *cmdError) Error() string {
	if e.msg != "" {
		return e.cmd + ": " + e.msg
	}
	return e.cmd + ": " + e.err.Error()
}
func (e *cmdError) Unwrap() error { return e.err }

func parseSecurity(s string) domain.WiFiSecurity {
	u := strings.ToUpper(s)
	switch {
	case u == "" || strings.Contains(u, "NONE") || u == "OPEN" || u == "--":
		return domain.WiFiOpen
	case strings.Contains(u, "WPA3") || strings.Contains(u, "SAE"):
		return domain.WiFiWPA3
	case strings.Contains(u, "WPA2") || strings.Contains(u, "RSN"):
		return domain.WiFiWPA2
	case strings.Contains(u, "WPA"):
		return domain.WiFiWPA
	case strings.Contains(u, "WEP"):
		return domain.WiFiWEP
	default:
		return domain.WiFiUnknown
	}
}
