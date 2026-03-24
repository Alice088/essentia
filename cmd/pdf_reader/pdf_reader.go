package main

import (
	"Alice088/essentia/pkg/pdf_reader"
	"Alice088/essentia/pkg/size"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ledongthuc/pdf"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			_ = json.NewEncoder(os.Stdout).Encode(pdf_reader.ReadResponse{
				Error: fmt.Sprintf("panic: %v", r),
			})
		}
	}()

	if len(os.Args) < 2 {
		_ = json.NewEncoder(os.Stdout).Encode(pdf_reader.ReadResponse{
			Error: "no file specified",
		})
		return
	}

	path := os.Args[1]

	fi, err := os.Lstat(path)
	if err != nil {
		writeErr(err)
		return
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		writeErr(fmt.Errorf("symlinks not allowed"))
		return
	}

	if fi.Size() > size.MB5 {
		writeErr(fmt.Errorf("file too large"))
		return
	}

	f, r, err := pdf.Open(path)
	if err != nil {
		writeErr(fmt.Errorf("failed to open pdf: %w", err))
		return
	}
	defer f.Close()

	var buf bytes.Buffer
	text, err := r.GetPlainText()
	if err != nil {
		writeErr(fmt.Errorf("failed to get pdf text: %w\n", err))
		return
	}

	_, err = io.CopyN(&buf, text, size.MB5)
	if err != nil && err != io.EOF {
		writeErr(fmt.Errorf("failed to read buffer: %w\n", err))
		return
	}

	content := buf.String()
	_ = json.NewEncoder(os.Stdout).Encode(pdf_reader.ReadResponse{
		Text: content,
		Metadata: pdf_reader.Metadata{
			Size: buf.Len(),
		},
	})
}

func writeErr(err error) {
	_ = json.NewEncoder(os.Stdout).Encode(pdf_reader.ReadResponse{
		Error: err.Error(),
	})
}
