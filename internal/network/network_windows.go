//go:build windows

package network

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type windowsManager struct {
	baseManager
}

func newPlatformManager(log logging.Logger) Manager {
	return &windowsManager{baseManager{log: log}}
}

func (m *windowsManager) ScanWiFi(ctx context.Context, iface string) ([]domain.WiFiNetwork, error) {
	out, err := runCommand(ctx, "netsh", "wlan", "show", "networks", "mode=bssid")
	if err != nil {
		return nil, err
	}
	return parseNetsh(out), nil
}

func (m *windowsManager) ConnectWiFi(ctx context.Context, iface, ssid, passphrase string) error {
	if ssid == "" {
		return fmt.Errorf("ssid required")
	}
	profile := wlanProfileXML(ssid, passphrase)
	tmp, err := os.CreateTemp("", "print-wifi-*.xml")
	if err != nil {
		return err
	}
	path := tmp.Name()
	defer os.Remove(path)
	if _, err := tmp.WriteString(profile); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	if _, err := runCommand(ctx, "netsh", "wlan", "add", "profile", "filename="+filepath.Clean(path)); err != nil {
		return err
	}
	_, err = runCommand(ctx, "netsh", "wlan", "connect", "name="+ssid, "ssid="+ssid)
	return err
}

func parseNetsh(out string) []domain.WiFiNetwork {
	var nets []domain.WiFiNetwork
	var cur *domain.WiFiNetwork

	flush := func() {
		if cur != nil && cur.SSID != "" {
			nets = append(nets, *cur)
		}
	}

	for _, raw := range strings.Split(out, "\n") {
		line := strings.TrimSpace(raw)
		key, val, ok := splitColon(line)
		if !ok {
			continue
		}
		switch {
		case strings.HasPrefix(key, "SSID ") && !strings.HasPrefix(key, "BSSID"):
			flush()
			n := domain.WiFiNetwork{SSID: val}
			cur = &n
		case cur == nil:
		case key == "Authentication":
			cur.Security = parseSecurity(val)
		case strings.HasPrefix(key, "Signal"):
			cur.Signal = parsePercent(val)
		case key == "Channel":
			cur.Channel = atoiSafe(val)
		case strings.HasPrefix(key, "BSSID"):
			if cur.BSSID == "" {
				cur.BSSID = val
			}
		}
	}
	flush()
	return nets
}

func splitColon(line string) (key, val string, ok bool) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", "", false
	}
	return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:]), true
}

func parsePercent(s string) int {
	s = strings.TrimSuffix(strings.TrimSpace(s), "%")
	return atoiSafe(s)
}

func atoiSafe(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	if n < 0 {
		return 0
	}
	return n
}

func wlanProfileXML(ssid, passphrase string) string {
	auth := "open"
	encryption := "none"
	security := ""
	if passphrase != "" {
		auth = "WPA2PSK"
		encryption = "AES"
		security = fmt.Sprintf(`
        <sharedKey>
          <keyType>passPhrase</keyType>
          <protected>false</protected>
          <keyMaterial>%s</keyMaterial>
        </sharedKey>`, xmlEscape(passphrase))
	}
	return fmt.Sprintf(`<?xml version="1.0"?>
<WLANProfile xmlns="http://www.microsoft.com/networking/WLAN/profile/v1">
  <name>%s</name>
  <SSIDConfig><SSID><name>%s</name></SSID></SSIDConfig>
  <connectionType>ESS</connectionType>
  <connectionMode>auto</connectionMode>
  <MSM><security>
      <authEncryption>
        <authentication>%s</authentication>
        <encryption>%s</encryption>
        <useOneX>false</useOneX>
      </authEncryption>%s
  </security></MSM>
</WLANProfile>`, xmlEscape(ssid), xmlEscape(ssid), auth, encryption, security)
}

func xmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;")
	return r.Replace(s)
}
