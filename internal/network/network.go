package network

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type Manager interface {
	Interfaces() ([]domain.NetworkInterface, error)
	ScanWiFi(ctx context.Context, iface string) ([]domain.WiFiNetwork, error)
	ConnectWiFi(ctx context.Context, iface, ssid, passphrase string) error
}

func New(log logging.Logger) Manager {
	return newPlatformManager(log)
}

type Connectivity struct {
	client *http.Client
}

func NewConnectivity() *Connectivity {
	return &Connectivity{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Connectivity) Online(ctx context.Context) bool {
	if dialCheck(ctx, "1.1.1.1:443") || dialCheck(ctx, "8.8.8.8:443") {
		return true
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, "https://cloudflare.com", nil)
	if err != nil {
		return false
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode < 500
}

func dialCheck(ctx context.Context, addr string) bool {
	d := net.Dialer{Timeout: 3 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
