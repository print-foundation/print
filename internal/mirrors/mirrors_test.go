package mirrors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeFetcher map[string][]byte

func (f fakeFetcher) Fetch(ctx context.Context, url string) ([]byte, error) {
	if b, ok := f[url]; ok {
		return b, nil
	}
	return nil, errNotFound
}

var errNotFound = &netErr{}

type netErr struct{}

func (e *netErr) Error() string { return "not found" }

func TestResolveCuratedFallback(t *testing.T) {
	r := NewResolver().WithFetcher(fakeFetcher{})
	ms, err := r.Resolve(context.Background(), DistroArch, "DE")
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	assert.Equal(t, "DE", ms[0].Country)
}

func TestResolveUnsupported(t *testing.T) {
	r := NewResolver().WithFetcher(fakeFetcher{})
	_, err := r.Resolve(context.Background(), DistroWindows, "US")
	assert.ErrorIs(t, err, ErrUnsupportedDistro)
}

func TestParseDebianList(t *testing.T) {
	sample := `Site: ftp.de.debian.org
Country: DE Germany
Archive-ftp: /debian/
Archive-http: /debian/
Sponsor: German mirror

Site: ftp.us.debian.org
Country: US United States
Archive-http: /debian/
`
	ms, err := parseDebianList([]byte(sample), "DE")
	require.NoError(t, err)
	require.Len(t, ms, 1)
	assert.Equal(t, "ftp.de.debian.org", ms[0].Host)
	assert.Equal(t, "https://ftp.de.debian.org/debian", ms[0].BaseURL)
}

func TestParseDebianListAllCountries(t *testing.T) {
	sample := "Site: a.debian.org\nCountry: DE Germany\nArchive-http: /debian/\n\nSite: b.debian.org\nCountry: US\nArchive-http: /debian/\n"
	ms, err := parseDebianList([]byte(sample), "")
	require.NoError(t, err)
	assert.Len(t, ms, 2)
}

func TestParseUbuntuMirrorsTxt(t *testing.T) {
	sample := "http://old.example.com/ubuntu/\nhttps://mirror.example.com/ubuntu/\n# comment\n"
	ms, err := parseUbuntuMirrorsTxt([]byte(sample), "")
	require.NoError(t, err)
	require.Len(t, ms, 1)
	assert.Equal(t, "mirror.example.com", ms[0].Host)
}

func TestParseFedoraMirrorlist(t *testing.T) {
	sample := "# fedora mirrorlist\nhttps://download.fedoraproject.org/pub/fedora/linux/releases/40/Everything/x86_64/os/\nmetalink: https://mirrors.fedoraproject.org/metalink?repo=fedora-40\n"
	ms, err := parseFedoraMirrorlist([]byte(sample), "")
	require.NoError(t, err)
	assert.Len(t, ms, 2)
}

func TestLiveSourceTakesPrecedence(t *testing.T) {
	f := fakeFetcher{
		"https://www.debian.org/mirror/list-full": []byte("Site: live.debian.org\nCountry: FR France\nArchive-http: /debian/\n"),
	}
	r := NewResolver().WithFetcher(f)
	ms, err := r.Resolve(context.Background(), DistroDebian, "FR")
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	assert.Equal(t, "live.debian.org", ms[0].Host)
}
