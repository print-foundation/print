package hashes

import (
	"bufio"
	"io"
	"path"
	"strings"
)

type ChecksumFile struct {
	Algorithm Algorithm
	Digests   map[string]string
}

func (c ChecksumFile) Lookup(name string) (string, bool) {
	if d, ok := c.Digests[name]; ok {
		return d, true
	}
	base := path.Base(name)
	d, ok := c.Digests[base]
	return d, ok
}

func ParseChecksumFile(r io.Reader, algo Algorithm) (ChecksumFile, error) {
	out := ChecksumFile{Algorithm: algo, Digests: map[string]string{}}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		digest := strings.ToLower(fields[0])
		name := strings.TrimSpace(strings.TrimPrefix(line, fields[0]))
		name = strings.TrimPrefix(name, "*")
		name = strings.TrimSpace(name)
		name = strings.TrimPrefix(name, "*") // binary-mode marker can trail the spaces too

		if out.Algorithm == "" {
			out.Algorithm = algorithmFromLength(len(digest))
		}
		out.Digests[name] = digest
	}
	if err := scanner.Err(); err != nil {
		return ChecksumFile{}, err
	}
	return out, nil
}

func algorithmFromLength(hexLen int) Algorithm {
	switch hexLen {
	case 40:
		return SHA1
	case 64:
		return SHA256
	case 128:
		return SHA512
	default:
		return ""
	}
}
