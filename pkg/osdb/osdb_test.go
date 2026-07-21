package osdb

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundledCatalogLoadsAndValidates(t *testing.T) {
	c, err := Bundled()
	require.NoError(t, err)
	require.NotEmpty(t, c.OperatingSystems)

	for _, os := range c.OperatingSystems {
		for _, rel := range os.Releases {
			hasVerification := !rel.Checksum.IsZero() || rel.ChecksumURL != ""
			assert.True(t, hasVerification, "release %s has no checksum source", rel.ID)
			assert.True(t, strings.HasPrefix(rel.URL, "https://"), rel.ID)
		}
	}
}

func TestLoadRejectsInsecureURL(t *testing.T) {
	doc := `{
		"schema_version": 1,
		"operating_systems": [
			{"id":"x","name":"X","releases":[
				{"id":"x-1","url":"http://example.com/x.iso"}
			]}
		]
	}`
	_, err := Load(strings.NewReader(doc))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "https")
}

func TestLoadRejectsWrongSchema(t *testing.T) {
	doc := `{"schema_version": 999, "operating_systems": []}`
	_, err := Load(strings.NewReader(doc))
	require.Error(t, err)
}

func TestLoadRejectsDuplicateReleaseID(t *testing.T) {
	doc := `{
		"schema_version": 1,
		"operating_systems": [
			{"id":"a","name":"A","releases":[{"id":"dup","url":"https://e.com/a.iso"}]},
			{"id":"b","name":"B","releases":[{"id":"dup","url":"https://e.com/b.iso"}]}
		]
	}`
	_, err := Load(strings.NewReader(doc))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestFindRelease(t *testing.T) {
	c, err := Bundled()
	require.NoError(t, err)

	_, rel, ok := c.FindRelease("ubuntu-24.04.1-desktop-amd64")
	require.True(t, ok)
	assert.Equal(t, "Desktop", rel.Edition)

	_, _, ok = c.FindRelease("does-not-exist")
	assert.False(t, ok)
}

func TestFilterByArch(t *testing.T) {
	c, err := Bundled()
	require.NoError(t, err)

	amd := c.FilterByArch("amd64")
	require.NotEmpty(t, amd)

	sparc := c.FilterByArch("sparc")
	assert.Empty(t, sparc)
}

func TestReleasesStableOrder(t *testing.T) {
	c, err := Bundled()
	require.NoError(t, err)
	a := c.Releases()
	b := c.Releases()
	assert.Equal(t, a, b)
}
