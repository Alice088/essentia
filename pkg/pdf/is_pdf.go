package pdf

import (
	"errors"
	"io"
)

func IsPDF(r io.Reader) error {
	buf := make([]byte, 5)
	if _, err := io.ReadFull(r, buf); err != nil {
		return errors.New("invalid pdf")
	}

	if string(buf) != "%PDF-" {
		return errors.New("invalid pdf signature")
	}

	return nil
}
