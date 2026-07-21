//go:build !linux && !darwin && !windows

package network

import (
	"context"
	"fmt"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type fallbackManager struct {
	baseManager
}

func newPlatformManager(log logging.Logger) Manager {
	return &fallbackManager{baseManager{log: log}}
}

func (m *fallbackManager) ScanWiFi(ctx context.Context, iface string) ([]domain.WiFiNetwork, error) {
	return nil, fmt.Errorf("%w: wifi scanning on this platform", domain.ErrUnsupported)
}

func (m *fallbackManager) ConnectWiFi(ctx context.Context, iface, ssid, passphrase string) error {
	return fmt.Errorf("%w: wifi management on this platform", domain.ErrUnsupported)
}
