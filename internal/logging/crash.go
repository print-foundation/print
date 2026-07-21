package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"
)

type CrashReport struct {
	Time    time.Time `json:"time"`
	Version string    `json:"version"`
	OS      string    `json:"os"`
	Arch    string    `json:"arch"`
	Reason  string    `json:"reason"`
	Stack   string    `json:"stack"`
}

type CrashReporter struct {
	dir     string
	version string
	log     Logger
}

func NewCrashReporter(dir, version string, log Logger) (*CrashReporter, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create crash dir: %w", err)
	}
	return &CrashReporter{dir: dir, version: version, log: log}, nil
}

func (r *CrashReporter) Recover() {
	if v := recover(); v != nil {
		report := CrashReport{
			Time:    time.Now().UTC(),
			Version: r.version,
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
			Reason:  fmt.Sprint(v),
			Stack:   string(debug.Stack()),
		}
		if path, err := r.write(report); err != nil {
			r.log.Error("failed to write crash report", "error", err)
		} else {
			r.log.Error("crash report written", "path", path)
		}
		panic(v)
	}
}

func (r *CrashReporter) write(report CrashReport) (string, error) {
	name := fmt.Sprintf("crash-%s.json", report.Time.Format("20060102-150405"))
	path := filepath.Join(r.dir, name)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func (r *CrashReporter) List() ([]string, error) {
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) == ".json" {
			paths = append(paths, filepath.Join(r.dir, e.Name()))
		}
	}
	for i := 0; i < len(paths); i++ {
		for j := i + 1; j < len(paths); j++ {
			if paths[j] > paths[i] {
				paths[i], paths[j] = paths[j], paths[i]
			}
		}
	}
	return paths, nil
}
