package builder

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/print-foundation/print/internal/mirrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFakeTool(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	tool := filepath.Join(dir, name)
	script := "#!/bin/sh\nwhile [ \"$#\" -gt 0 ]; do\n  case \"$1\" in\n    -o) echo iso > \"$2\"; shift 2;;\n    *) shift;;\n  esac\ndone\n"
	require.NoError(t, os.WriteFile(tool, []byte(script), 0o755))
	return tool
}

func testClient(srv *httptest.Server) *http.Client {
	c := srv.Client()
	c.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	return c
}

func TestBuildAssemblesISOWithConfig(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ISO assembly requires Linux tools")
	}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch filepath.Base(r.URL.Path) {
		case "linux", "vmlinuz":
			w.Write([]byte("KERNEL-BYTES")) //nolint:errcheck
		case "initrd.gz", "initrd":
			w.Write([]byte("INITRD-BYTES")) //nolint:errcheck
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	old := netbootAssets
	defer func() { netbootAssets = old }()
	netbootAssets = map[mirrors.Distro]map[Arch]Asset{
		mirrors.DistroDebian: {ArchAmd64: {Kernel: "/linux", Initrd: "/initrd.gz"}},
	}

	work := t.TempDir()
	tool := writeFakeTool(t, "grub-mkrescue.cmd")
	out := filepath.Join(t.TempDir(), "print-debian.iso")

	b := New(WithHTTPClient(testClient(srv)))
	res, err := b.Build(context.Background(), Spec{
		Distro:  mirrors.DistroDebian,
		Arch:    ArchAmd64,
		Mirror:  mirrors.Mirror{BaseURL: srv.URL},
		Output:  out,
		WorkDir: work,
		Tools:   ToolPaths{GrubMkRescue: tool},
	}, []byte("CLIENT-BINARY"))
	require.NoError(t, err)
	assert.Equal(t, out, res.ISO)
	assert.FileExists(t, out)

	cfgRaw, err := os.ReadFile(filepath.Join(work, "print-client.json"))
	require.NoError(t, err)
	var cfg ClientConfig
	require.NoError(t, json.Unmarshal(cfgRaw, &cfg))
	assert.Equal(t, mirrors.DistroDebian, cfg.Distro)
	assert.Equal(t, srv.URL, cfg.Mirror)
	assert.FileExists(t, filepath.Join(work, "print-client"))
}

func TestBuildRejectsUnsupported(t *testing.T) {
	b := New()
	_, err := b.Build(context.Background(), Spec{
		Distro: mirrors.DistroWindows,
		Arch:   ArchAmd64,
		Output: "x.iso",
	}, nil)
	assert.ErrorIs(t, err, errUnsupported)
}

func TestBuildNeedsISOTool(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ISO assembly requires Linux tools")
	}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("data")) //nolint:errcheck
	}))
	defer srv.Close()
	old := netbootAssets
	defer func() { netbootAssets = old }()
	netbootAssets = map[mirrors.Distro]map[Arch]Asset{
		mirrors.DistroDebian: {ArchAmd64: {Kernel: "/linux", Initrd: "/initrd"}},
	}
	b := New(WithHTTPClient(testClient(srv)))
	_, err := b.Build(context.Background(), Spec{
		Distro:  mirrors.DistroDebian,
		Arch:    ArchAmd64,
		Mirror:  mirrors.Mirror{BaseURL: srv.URL},
		Output:  filepath.Join(t.TempDir(), "o.iso"),
		WorkDir: t.TempDir(),
		Tools:   ToolPaths{GrubMkRescue: "definitely-not-a-real-tool-xyz"},
	}, nil)
	assert.Error(t, err)
}
