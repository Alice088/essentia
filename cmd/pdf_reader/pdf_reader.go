package main

import (
	"Alice088/pdf-summarize/pkg/pdf_reader"
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ledongthuc/pdf"
)

func main() {
	if len(os.Args) < 2 {
		_ = json.NewEncoder(os.Stdout).Encode(pdf_reader.ReadResponse{
			Error: "no file specified",
		})
		return
	}

	path := os.Args[1]

	f, r, err := pdf.Open(path)
	if err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(pdf_reader.ReadResponse{
			Error: fmt.Sprintf("failed to open pdf: %s", err.Error()),
		})
		return
	}
	defer f.Close()

	var buf bytes.Buffer
	text, err := r.GetPlainText()
	if err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(pdf_reader.ReadResponse{
			Error: fmt.Sprintf("failed to get pdf text: %s\n", err.Error()),
		})
		return
	}

	_, err = buf.ReadFrom(text)
	if err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(pdf_reader.ReadResponse{
			Error: fmt.Sprintf("failed to read buffer: %s\n", err.Error()),
		})
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
