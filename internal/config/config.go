package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/print-foundation/print/internal/logging"
)

type Theme string

const (
	ThemeDark   Theme = "dark"
	ThemeLight  Theme = "light"
	ThemeSystem Theme = "system"
)

type Config struct {
	Theme Theme `json:"theme"`

	Language string `json:"language"`

	LogLevel logging.Level `json:"log_level"`

	DownloadDir string `json:"download_dir"`

	MaxParallelDownloads int `json:"max_parallel_downloads"`

	VerifyDownloads bool `json:"verify_downloads"`

	AutoUpdate bool `json:"auto_update"`

	CatalogURL string `json:"catalog_url"`

	EnabledPlugins []string `json:"enabled_plugins"`
}

func Default() Config {
	return Config{
		Theme:                ThemeSystem,
		Language:             "en",
		LogLevel:             logging.LevelInfo,
		MaxParallelDownloads: 4,
		VerifyDownloads:      true,
		AutoUpdate:           true,
	}
}

func (c *Config) Validate() error {
	switch c.Theme {
	case ThemeDark, ThemeLight, ThemeSystem:
	default:
		c.Theme = ThemeSystem
	}
	if c.Language == "" {
		c.Language = "en"
	}
	if c.MaxParallelDownloads < 1 {
		c.MaxParallelDownloads = 1
	}
	if c.MaxParallelDownloads > 16 {
		c.MaxParallelDownloads = 16
	}
	switch c.LogLevel {
	case logging.LevelDebug, logging.LevelInfo, logging.LevelWarn, logging.LevelError:
	default:
		c.LogLevel = logging.LevelInfo
	}
	return nil
}

type Store struct {
	mu   sync.RWMutex
	path string
	cfg  Config
}

func NewStore(path string) *Store {
	return &Store{path: path, cfg: Default()}
}

func (s *Store) Load() (Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		s.cfg = Default()
		return s.cfg, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	cfg := Default()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	s.cfg = cfg
	return cfg, nil
}

func (s *Store) Save(cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}
	s.cfg = cfg
	return nil
}

func (s *Store) Current() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}
