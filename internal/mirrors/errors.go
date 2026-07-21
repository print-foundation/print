package mirrors

import (
	"fmt"
	"io"
)

var ErrUnsupportedDistro = fmt.Errorf("unsupported distro")

var ErrNoMirror = fmt.Errorf("no mirror")

func readAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, 16*1024*1024))
}
