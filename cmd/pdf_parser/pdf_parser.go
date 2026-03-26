package main

import (
	errs "Alice088/essentia/pkg/errors"
	"Alice088/essentia/pkg/pdf_parser"
	"Alice088/essentia/pkg/real_size"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			errT := errs.ErrUnknown
			if err, ok := r.(error); ok {
				errT = classifyOpenPDFError(err)
			}
			writeErr(errT, fmt.Errorf("panic: %v", r))
		}
	}()

	if len(os.Args) < 2 {
		writeErr(errs.ErrOpen, fmt.Errorf("no file specified"))
		return
	}

	path := os.Args[1]

	fi, err := os.Lstat(path)
	if err != nil {
		writeErr(errs.ErrOpen, err)
		return
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		writeErr(errs.ErrOpen, fmt.Errorf("symlinks not allowed"))
		return
	}

	if fi.Size() > real_size.MB5 {
		writeErr(errs.ErrOpen, fmt.Errorf("file too large"))
		return
	}

	f, r, err := pdf.Open(path)
	if err != nil {
		writeErr(classifyOpenPDFError(err), fmt.Errorf("failed to open pdf: %w", err))
		return
	}
	defer f.Close()

	var buf bytes.Buffer
	text, err := r.GetPlainText()
	if err != nil {
		writeErr(classifyExtractError(err), fmt.Errorf("failed to get pdf text: %w", err))
		return
	}

	_, err = io.CopyN(&buf, text, real_size.MB5)
	if err != nil && err != io.EOF {
		writeErr(errs.ErrExtract, fmt.Errorf("failed to read buffer: %w", err))
		return
	}

	if buf.Len() == 0 {
		writeErr(errs.ErrEmpty, errors.New("pdf is empty"))
		return
	}

	content := buf.String()
	_ = json.NewEncoder(os.Stdout).Encode(pdf_parser.ReadResponse{
		Text: content,
		Metadata: pdf_parser.Metadata{
			Size: buf.Len(),
		},
	})
}

func writeErr(code errs.PipelineError, err error) {
	_ = json.NewEncoder(os.Stdout).Encode(pdf_parser.ReadResponse{
		Error:     err.Error(),
		ErrorCode: string(code),
	})
}

func classifyOpenPDFError(err error) errs.PipelineError {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "encrypted") || strings.Contains(msg, "password") {
		return errs.ErrEncrypted
	}

	if strings.Contains(msg, "malformed") || strings.Contains(msg, "corrupt") || strings.Contains(msg, "invalid") {
		return errs.ErrCorrupted
	}

	return errs.ErrOpen
}

func classifyExtractError(err error) errs.PipelineError {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "encrypted") || strings.Contains(msg, "password") {
		return errs.ErrEncrypted
	}

	if strings.Contains(msg, "malformed") || strings.Contains(msg, "corrupt") || strings.Contains(msg, "invalid") {
		return errs.ErrCorrupted
	}

	return errs.ErrExtract
}
