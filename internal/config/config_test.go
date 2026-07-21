package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/print-foundation/print/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultIsValid(t *testing.T) {
	cfg := Default()
	require.NoError(t, cfg.Validate())
	assert.True(t, cfg.VerifyDownloads)
	assert.Equal(t, ThemeSystem, cfg.Theme)
}

func TestValidateRepairsBadValues(t *testing.T) {
	cfg := Config{
		Theme:                "neon",
		Language:             "",
		LogLevel:             "loud",
		MaxParallelDownloads: 999,
	}
	require.NoError(t, cfg.Validate())
	assert.Equal(t, ThemeSystem, cfg.Theme)
	assert.Equal(t, "en", cfg.Language)
	assert.Equal(t, logging.LevelInfo, cfg.LogLevel)
	assert.Equal(t, 16, cfg.MaxParallelDownloads)

	cfg.MaxParallelDownloads = 0
	require.NoError(t, cfg.Validate())
	assert.Equal(t, 1, cfg.MaxParallelDownloads)
}

func TestStoreLoadMissingReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	s := NewStore(path)
	cfg, err := s.Load()
	require.NoError(t, err)
	assert.Equal(t, Default(), cfg)
}

func TestStoreSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	s := NewStore(path)

	cfg := Default()
	cfg.Theme = ThemeDark
	cfg.Language = "pt-BR"
	cfg.EnabledPlugins = []string{"example"}
	require.NoError(t, s.Save(cfg))

	s2 := NewStore(path)
	loaded, err := s2.Load()
	require.NoError(t, err)
	assert.Equal(t, ThemeDark, loaded.Theme)
	assert.Equal(t, "pt-BR", loaded.Language)
	assert.Equal(t, []string{"example"}, loaded.EnabledPlugins)
}

func TestStoreSaveIsAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	s := NewStore(path)
	require.NoError(t, s.Save(Default()))

	_, err := os.Stat(path + ".tmp")
	assert.True(t, os.IsNotExist(err))
}

func TestResolvePaths(t *testing.T) {
	p, err := ResolvePaths()
	require.NoError(t, err)
	assert.Contains(t, p.ConfigFile, appName)
	assert.Contains(t, p.LogDir, "logs")
	assert.Contains(t, p.CrashDir, "crashes")
}
