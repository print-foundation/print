//go:build darwin

package network

import (
	"context"
	"fmt"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type darwinManager struct {
	baseManager
}

func newPlatformManager(log logging.Logger) Manager {
	return &darwinManager{baseManager{log: log}}
}

func (m *darwinManager) ScanWiFi(ctx context.Context, iface string) ([]domain.WiFiNetwork, error) {
	return nil, fmt.Errorf("%w: wifi scanning on macOS requires system Wi-Fi settings", domain.ErrUnsupported)
}

func (m *darwinManager) ConnectWiFi(ctx context.Context, iface, ssid, passphrase string) error {
	if ssid == "" {
		return fmt.Errorf("ssid required")
	}
	if iface == "" {
		iface = "en0" // the built-in wifi device on most macs
	}
	args := []string{"-setairportnetwork", iface, ssid}
	if passphrase != "" {
		args = append(args, passphrase)
	}
	_, err := runCommand(ctx, "networksetup", args...)
	return err
}
