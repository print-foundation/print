package network

import (
	"context"
	"testing"
	"time"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSecurity(t *testing.T) {
	cases := map[string]domain.WiFiSecurity{
		"":            domain.WiFiOpen,
		"--":          domain.WiFiOpen,
		"WPA3":        domain.WiFiWPA3,
		"WPA2":        domain.WiFiWPA2,
		"WPA1 WPA2":   domain.WiFiWPA2,
		"WPA":         domain.WiFiWPA,
		"WEP":         domain.WiFiWEP,
		"RSN-PSK-CCM": domain.WiFiWPA2,
		"weirdstuff":  domain.WiFiUnknown,
	}
	for in, want := range cases {
		assert.Equal(t, want, parseSecurity(in), in)
	}
}

func TestNameLooksWireless(t *testing.T) {
	assert.True(t, nameLooksWireless("wlan0"))
	assert.True(t, nameLooksWireless("Wi-Fi"))
	assert.False(t, nameLooksWireless("eth0"))
}

func TestBaseInterfaces(t *testing.T) {
	b := baseManager{log: logging.NopLogger()}
	ifaces, err := b.Interfaces()
	require.NoError(t, err)
	for _, ni := range ifaces {
		assert.NotEmpty(t, ni.Name)
	}
}

func TestConnectivityOfflineFast(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c := NewConnectivity()
	assert.False(t, c.Online(ctx))
}

func TestConnectivityRespectsTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	c := NewConnectivity()
	start := time.Now()
	_ = c.Online(ctx)
	assert.Less(t, time.Since(start), 8*time.Second)
}
