package main

import (
	errs "Alice088/essentia/pkg/errors"
	"Alice088/essentia/pkg/pdf_reader"
	"Alice088/essentia/pkg/size"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			writeErr(errs.ParsingErrUnknown, fmt.Errorf("panic: %v", r))
		}
	}()

	if len(os.Args) < 2 {
		writeErr(errs.ParsingErrOpen, fmt.Errorf("no file specified"))
		return
	}

	path := os.Args[1]

	fi, err := os.Lstat(path)
	if err != nil {
		writeErr(errs.ParsingErrOpen, err)
		return
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		writeErr(errs.ParsingErrOpen, fmt.Errorf("symlinks not allowed"))
		return
	}

	if fi.Size() > size.MB5 {
		writeErr(errs.ParsingErrOpen, fmt.Errorf("file too large"))
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

	_, err = io.CopyN(&buf, text, size.MB5)
	if err != nil && err != io.EOF {
		writeErr(errs.ParsingErrExtract, fmt.Errorf("failed to read buffer: %w", err))
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

func writeErr(code errs.ParsingErrorCode, err error) {
	_ = json.NewEncoder(os.Stdout).Encode(pdf_reader.ReadResponse{
		Error:     err.Error(),
		ErrorCode: string(code),
	})
}

func classifyOpenPDFError(err error) errs.ParsingErrorCode {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "encrypted") || strings.Contains(msg, "password") {
		return errs.ParsingErrEncrypted
	}

	if strings.Contains(msg, "malformed") || strings.Contains(msg, "corrupt") || strings.Contains(msg, "invalid") {
		return errs.ParsingErrCorrupted
	}

	return errs.ParsingErrOpen
}

func classifyExtractError(err error) errs.ParsingErrorCode {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "encrypted") || strings.Contains(msg, "password") {
		return errs.ParsingErrEncrypted
	}

	if strings.Contains(msg, "malformed") || strings.Contains(msg, "corrupt") || strings.Contains(msg, "invalid") {
		return errs.ParsingErrCorrupted
	}

	return errs.ParsingErrExtract
}
