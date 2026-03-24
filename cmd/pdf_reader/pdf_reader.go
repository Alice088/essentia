package main

import (
	"Alice088/essentia/pkg/pdf_reader"
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ledongthuc/pdf"
)

//TODO
//1) Библиотека устаревшая  - не поддерживает PDF 1.7
//2) Panic вместо graceful error - небезопасно
//============================================
//1. PDF с 2003 объектами - принят
//    Риск: перегрузка памяти парсера
//    Нужно: лимит на количество объектов (max 1000)
//2. PDF с рекурсивными ссылками - принят
//    Риск: бесконечная рекурсия, stack overflow
//    Нужно: лимит на глубину рекурсии (max 50)
//3. PDF с JavaScript - принят
//    Риск: XSS при просмотре, выполнение кода
//    Нужно: запрет JavaScript действий
//4. PDF со сжатыми данными - принят
//    Риск: ZIP bomb (10KB → 10MB при распаковке)
//    Нужно: лимит на распакованный размер

func main() {
	defer func() {
		if r := recover(); r != nil {
			_ = json.NewEncoder(os.Stdout).Encode(pdf_reader.ReadResponse{
				Error: "failed to parse pdf: " + r.(error).Error(),
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
