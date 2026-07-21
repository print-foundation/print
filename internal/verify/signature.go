package verify

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
)

var ErrNoTrustedKey = errors.New("no trusted key available")

type PGPVerifier struct {
	keyring openpgp.EntityList
}

func NewPGPVerifier(keys ...io.Reader) (*PGPVerifier, error) {
	v := &PGPVerifier{}
	for _, r := range keys {
		if err := v.addKey(r); err != nil {
			return nil, err
		}
	}
	return v, nil
}

func LoadPGPKeyDir(dir string) (*PGPVerifier, error) {
	v := &PGPVerifier{}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return v, nil
	}
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		switch filepath.Ext(e.Name()) {
		case ".asc", ".gpg", ".key", ".pub":
		default:
			continue
		}
		f, err := os.Open(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		err = v.addKey(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("load key %s: %w", e.Name(), err)
		}
	}
	return v, nil
}

func (v *PGPVerifier) addKey(r io.Reader) error {
	data, err := io.ReadAll(io.LimitReader(r, 4*1024*1024))
	if err != nil {
		return err
	}
	if el, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(data)); err == nil {
		v.keyring = append(v.keyring, el...)
		return nil
	}
	el, err := openpgp.ReadKeyRing(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("parse key: %w", err)
	}
	v.keyring = append(v.keyring, el...)
	return nil
}

func (v *PGPVerifier) Verify(_ context.Context, data, sig []byte) error {
	if len(v.keyring) == 0 {
		return ErrNoTrustedKey
	}

	sigReader := io.Reader(bytes.NewReader(sig))
	if block, err := armor.Decode(bytes.NewReader(sig)); err == nil {
		sigReader = block.Body
	}

	_, err := openpgp.CheckDetachedSignature(v.keyring, bytes.NewReader(data), sigReader, nil)
	if err != nil {
		return err
	}
	return nil
}

func (v *PGPVerifier) KeyCount() int {
	return len(v.keyring)
}
