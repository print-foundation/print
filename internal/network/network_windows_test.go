package network

import (
	"testing"

	"github.com/print-foundation/print/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNetsh(t *testing.T) {
	sample := `Interface name : Wi-Fi
There are 2 networks currently visible.

SSID 1 : HomeNetwork
    Network type            : Infrastructure
    Authentication          : WPA2-Personal
    Encryption              : CCMP
    BSSID 1                 : aa:bb:cc:dd:ee:ff
         Signal             : 92%
         Radio type         : 802.11ac
         Channel            : 36

SSID 2 : CoffeeShop
    Network type            : Infrastructure
    Authentication          : Open
    Encryption              : None
    BSSID 1                 : 11:22:33:44:55:66
         Signal             : 47%
         Channel            : 6
`
	nets := parseNetsh(sample)
	require.Len(t, nets, 2)

	assert.Equal(t, "HomeNetwork", nets[0].SSID)
	assert.Equal(t, domain.WiFiWPA2, nets[0].Security)
	assert.Equal(t, 92, nets[0].Signal)
	assert.Equal(t, 36, nets[0].Channel)
	assert.Equal(t, "aa:bb:cc:dd:ee:ff", nets[0].BSSID)

	assert.Equal(t, "CoffeeShop", nets[1].SSID)
	assert.Equal(t, domain.WiFiOpen, nets[1].Security)
	assert.Equal(t, 47, nets[1].Signal)
}

func TestWLANProfileXMLEscapes(t *testing.T) {
	xml := wlanProfileXML("Guest & Co", "p@ss<word>")
	assert.Contains(t, xml, "Guest &amp; Co")
	assert.Contains(t, xml, "p@ss&lt;word&gt;")
	assert.Contains(t, xml, "WPA2PSK")

	open := wlanProfileXML("Open", "")
	assert.Contains(t, open, "<authentication>open</authentication>")
	assert.NotContains(t, open, "sharedKey")
}
