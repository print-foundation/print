package config

import (
	"os"
	"path/filepath"
)

type Paths struct {
	ConfigFile string
	LogDir     string
	CrashDir   string
	CacheDir   string
	PluginDir  string
}

const appName = "print"

func ResolvePaths() (Paths, error) {
	cfgRoot, err := os.UserConfigDir()
	if err != nil {
		return Paths{}, err
	}
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return Paths{}, err
	}

	base := filepath.Join(cfgRoot, appName)
	cache := filepath.Join(cacheRoot, appName)

	return Paths{
		ConfigFile: filepath.Join(base, "config.json"),
		LogDir:     filepath.Join(base, "logs"),
		CrashDir:   filepath.Join(base, "crashes"),
		CacheDir:   cache,
		PluginDir:  filepath.Join(base, "plugins"),
	}, nil
}
