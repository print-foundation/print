package main

import (
	"encoding/json"
	"testing"

	"github.com/print-foundation/print/internal/builder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSeedArchIsValidConfig(t *testing.T) {
	seed, err := generateSeed(builder.ClientConfig{Distro: "arch", Mirror: "https://mirror.example.com/archlinux"})
	require.NoError(t, err)
	var cfg map[string]any
	require.NoError(t, json.Unmarshal([]byte(seed), &cfg))
	assert.Contains(t, cfg, "mirror_regions")
}

func TestGenerateSeedRejectsUnknown(t *testing.T) {
	_, err := generateSeed(builder.ClientConfig{Distro: "windows"})
	assert.Error(t, err)
}

func TestGenerateSeedDebianMentionsMirror(t *testing.T) {
	seed, err := generateSeed(builder.ClientConfig{Distro: "debian", Mirror: "https://deb.debian.org/debian"})
	require.NoError(t, err)
	assert.Contains(t, seed, "deb.debian.org")
}
