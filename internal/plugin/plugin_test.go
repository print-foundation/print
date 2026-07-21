package plugin

import (
	"context"
	"testing"

	"github.com/print-foundation/print/pkg/osdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countingSource struct{ released bool }

func (c *countingSource) Sources(ctx context.Context) ([]osdb.Release, error) {
	return []osdb.Release{{ID: "plugin/extra", URL: "https://example/extra.iso", Format: "iso"}}, nil
}

type abortHook struct{ NoOpHook }

func (abortHook) OnPhase(ctx context.Context, phase Phase, req Request) error {
	if phase == PhaseVerified {
		return assertErr
	}
	return nil
}

var assertErr = &pluginErr{}

type pluginErr struct{}

func (pluginErr) Error() string { return "aborted by plugin" }

func TestRegisterAndList(t *testing.T) {
	err := Register(Extension{Meta: Meta{ID: "demo", Name: "Demo", Version: "1.0"}})
	require.NoError(t, err)
	assert.Len(t, List(), 1)

	err = Register(Extension{Meta: Meta{ID: "demo"}})
	assert.Error(t, err)

	assert.Error(t, Register(Extension{}))
}

func TestRunHooksStopsOnError(t *testing.T) {
	require.NoError(t, Register(Extension{Meta: Meta{ID: "abort", Name: "Abort"}, Hook: abortHook{}}))
	err := RunHooks(context.Background(), PhaseVerified, Request{ReleaseID: "x"})
	assert.Error(t, err)
}

func TestAllSourcesMerges(t *testing.T) {
	src := &countingSource{}
	require.NoError(t, Register(Extension{Meta: Meta{ID: "src", Name: "Src"}, Sources: src}))
	sources := AllSources()
	require.Len(t, sources, 1)
	rels, err := sources[0].Sources(context.Background())
	require.NoError(t, err)
	require.Len(t, rels, 1)
	assert.Equal(t, "plugin/extra", rels[0].ID)
}
