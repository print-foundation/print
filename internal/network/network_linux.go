//go:build linux

package network

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type linuxManager struct {
	baseManager
}

func newPlatformManager(log logging.Logger) Manager {
	return &linuxManager{baseManager{log: log}}
}

func (m *linuxManager) ScanWiFi(ctx context.Context, iface string) ([]domain.WiFiNetwork, error) {
	args := []string{"-t", "-f", "SSID,BSSID,SIGNAL,SECURITY,CHAN", "device", "wifi", "list", "--rescan", "yes"}
	if iface != "" {
		args = append(args, "ifname", iface)
	}
	out, err := runCommand(ctx, "nmcli", args...)
	if err != nil {
		return nil, err
	}
	return parseNmcli(out), nil
}

func (m *linuxManager) ConnectWiFi(ctx context.Context, iface, ssid, passphrase string) error {
	if ssid == "" {
		return fmt.Errorf("ssid required")
	}
	args := []string{"device", "wifi", "connect", ssid}
	if passphrase != "" {
		args = append(args, "password", passphrase)
	}
	if iface != "" {
		args = append(args, "ifname", iface)
	}
	_, err := runCommand(ctx, "nmcli", args...)
	return err
}

func parseNmcli(out string) []domain.WiFiNetwork {
	var nets []domain.WiFiNetwork
	seen := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := splitEscaped(line, ':')
		if len(fields) < 5 {
			continue
		}
		ssid := fields[0]
		if ssid == "" || seen[ssid] {
			continue
		}
		seen[ssid] = true
		nets = append(nets, domain.WiFiNetwork{
			SSID:     ssid,
			BSSID:    strings.ReplaceAll(fields[1], "\\", ""),
			Signal:   atoiClamp(fields[2]),
			Security: parseSecurity(fields[3]),
			Channel:  atoi(fields[4]),
		})
	}
	return nets
}

func splitEscaped(s string, sep byte) []string {
	var out []string
	var cur strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			cur.WriteByte(s[i+1])
			i++
			continue
		}
		if s[i] == sep {
			out = append(out, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteByte(s[i])
	}
	out = append(out, cur.String())
	return out
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func atoiClamp(s string) int {
	n := atoi(s)
	if n < 0 {
		return 0
	}
	if n > 100 {
		return 100
	}
	return n
}
