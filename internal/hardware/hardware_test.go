package hardware

import (
	"context"
	"testing"
	"time"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchitectureIsKnown(t *testing.T) {
	b := baseDetector{log: logging.NopLogger()}
	assert.NotEqual(t, domain.ArchUnknown, b.architecture())
}

func TestNetworkInterfacesNoLoopback(t *testing.T) {
	b := baseDetector{log: logging.NopLogger()}
	for _, ni := range b.networkInterfaces() {
		assert.NotEqual(t, "lo", ni.Name)
	}
}

func TestLooksWireless(t *testing.T) {
	assert.True(t, looksWireless("wlan0"))
	assert.True(t, looksWireless("Wi-Fi"))
	assert.True(t, looksWireless("wlp3s0"))
	assert.False(t, looksWireless("eth0"))
	assert.False(t, looksWireless("Ethernet"))
}

func TestDetectRuns(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	d := New(logging.NopLogger())
	hw, err := d.Detect(ctx)
	require.NoError(t, err)

	assert.NotEqual(t, domain.ArchUnknown, hw.Architecture)
	assert.GreaterOrEqual(t, hw.CPU.Threads, 1, "should detect at least one thread")
}
