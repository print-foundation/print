package hardware

import (
	"context"
	"net"
	"runtime"
	"strings"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type Detector interface {
	Detect(ctx context.Context) (domain.Hardware, error)
}

func New(log logging.Logger) Detector {
	return newPlatformDetector(log)
}

type baseDetector struct {
	log logging.Logger
}

func (b baseDetector) architecture() domain.Architecture {
	return domain.NormalizeArch(runtime.GOARCH)
}

func (b baseDetector) networkInterfaces() []domain.NetworkInterface {
	ifaces, err := net.Interfaces()
	if err != nil {
		b.log.Warn("failed to list network interfaces", "error", err)
		return nil
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
			Wireless:     looksWireless(ifc.Name),
		}
		if addrs, err := ifc.Addrs(); err == nil {
			for _, a := range addrs {
				ni.Addresses = append(ni.Addresses, a.String())
			}
		}
		out = append(out, ni)
	}
	return out
}

func looksWireless(name string) bool {
	lower := strings.ToLower(name)
	for _, hint := range []string{"wlan", "wlp", "wl", "wi-fi", "wifi", "wireless", "airport"} {
		if strings.Contains(lower, hint) {
			return true
		}
	}
	return false
}
