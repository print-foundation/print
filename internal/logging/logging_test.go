package logging

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	log := New(Options{Level: LevelWarn, Writer: &buf})

	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")

	out := buf.String()
	assert.NotContains(t, out, "debug message")
	assert.NotContains(t, out, "info message")
	assert.Contains(t, out, "warn message")
	assert.Contains(t, out, "error message")
}

func TestLoggerWith(t *testing.T) {
	var buf bytes.Buffer
	log := New(Options{Level: LevelInfo, Writer: &buf, JSON: true}).With("component", "test")
	log.Info("hello")

	var record map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &record))
	assert.Equal(t, "test", record["component"])
	assert.Equal(t, "hello", record["msg"])
}

func TestTee(t *testing.T) {
	var a, b bytes.Buffer
	log := NewTee(
		New(Options{Level: LevelInfo, Writer: &a}),
		New(Options{Level: LevelInfo, Writer: &b}),
	)
	log.Info("fanned out")
	assert.Contains(t, a.String(), "fanned out")
	assert.Contains(t, b.String(), "fanned out")
}

func TestCrashReporter(t *testing.T) {
	dir := t.TempDir()
	r, err := NewCrashReporter(dir, "v1.2.3", NopLogger())
	require.NoError(t, err)

	func() {
		defer func() {
			_ = recover()
		}()
		defer r.Recover()
		panic("boom")
	}()

	paths, err := r.List()
	require.NoError(t, err)
	require.Len(t, paths, 1)

	data, err := os.ReadFile(paths[0])
	require.NoError(t, err)
	var report CrashReport
	require.NoError(t, json.Unmarshal(data, &report))
	assert.Equal(t, "boom", report.Reason)
	assert.Equal(t, "v1.2.3", report.Version)
	assert.True(t, strings.HasPrefix(filepath.Base(paths[0]), "crash-"))
}
