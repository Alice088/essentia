package pdf_parser

import (
	"fmt"
	"os"
)

type TMP struct {
	F        *os.File
	FileSize *int64
}

func (tmp *TMP) Path() string {
	return tmp.F.Name()
}

func (tmp *TMP) Size() (int64, error) {
	if tmp.FileSize != nil {
		return *tmp.FileSize, nil
	}

	info, err := tmp.F.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to get size: %w", err)
	}

	tmp.FileSize = new(info.Size())
	return *tmp.FileSize, nil
}

func (r *Parser) CreateTMP() error {
	f, err := os.CreateTemp("", "*.pdf")
	if err != nil {
		return fmt.Errorf("failed to create tmp file: %w", err)
	}
	r.TMP = TMP{f, nil}
	return nil
}
