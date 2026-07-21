package domain

import "net"

type NetworkInterface struct {
	Name         string
	HardwareAddr string
	Wireless     bool
	Up           bool
	Addresses    []string
}

type WiFiNetwork struct {
	SSID     string
	BSSID    string
	Signal   int // signal quality as a percentage, 0-100
	Security WiFiSecurity
	Channel  int
}

type WiFiSecurity string

const (
	WiFiOpen    WiFiSecurity = "open"
	WiFiWEP     WiFiSecurity = "wep"
	WiFiWPA     WiFiSecurity = "wpa"
	WiFiWPA2    WiFiSecurity = "wpa2"
	WiFiWPA3    WiFiSecurity = "wpa3"
	WiFiUnknown WiFiSecurity = "unknown"
)

type NetworkConfig struct {
	Interface string
	DHCP      bool

	Address net.IP
	Netmask net.IPMask
	Gateway net.IP
	DNS     []net.IP
}
