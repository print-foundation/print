package network

import (
	"testing"

	"github.com/print-foundation/print/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNmcli(t *testing.T) {
	sample := `HomeNetwork:AA\:BB\:CC\:DD\:EE\:FF:88:WPA2:36
CoffeeShop:11\:22\:33\:44\:55\:66:52::6
HomeNetwork:AA\:BB\:CC\:DD\:EE\:00:70:WPA2:40
`
	nets := parseNmcli(sample)
	require.Len(t, nets, 2, "duplicate ssid should be deduped")

	assert.Equal(t, "HomeNetwork", nets[0].SSID)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", nets[0].BSSID)
	assert.Equal(t, 88, nets[0].Signal)
	assert.Equal(t, domain.WiFiWPA2, nets[0].Security)
	assert.Equal(t, 36, nets[0].Channel)

	assert.Equal(t, "CoffeeShop", nets[1].SSID)
	assert.Equal(t, domain.WiFiOpen, nets[1].Security)
}

func TestSplitEscaped(t *testing.T) {
	got := splitEscaped(`a:b\:c:d`, ':')
	assert.Equal(t, []string{"a", "b:c", "d"}, got)
}

func TestAtoiClamp(t *testing.T) {
	assert.Equal(t, 0, atoiClamp("-5"))
	assert.Equal(t, 100, atoiClamp("150"))
	assert.Equal(t, 42, atoiClamp("42"))
}
