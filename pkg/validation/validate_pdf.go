package validation

import (
	errs "Alice088/essentia/pkg/errors"
	"Alice088/essentia/pkg/size"
	"bytes"
	"errors"
	"io"
	"net/http"
)

const pdfHeaderSize = 5

type Result struct {
	Error    error
	Metadata SafeMetadata
}

type SafeMetadata struct {
	Size   int64
	Reader io.Reader
}

type PDFInput struct {
	Reader io.Reader
	Size   int64
}

func PDF(input PDFInput) Result {
	res := Result{
		Error: nil,
	}

	if input.Size > size.MB5 {
		return Result{
			Error: &errs.PDFError{
				Err:  errors.New("file too large"),
				Code: errs.ErrTooLarge,
			},
		}
	}

	header := make([]byte, pdfHeaderSize)
	if _, err := io.ReadFull(input.Reader, header); err != nil {
		return Result{
			Error: &errs.PDFError{
				Err:  errors.New("invalid request body"),
				Code: errs.ErrBadInput,
			},
		}
	}

	if string(header) != "%PDF-" {
		return Result{
			Error: &errs.PDFError{
				Err:  errors.New("not a pdf file"),
				Code: errs.ErrNotPDF,
			},
		}
	}

	body, err := io.ReadAll(input.Reader)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return Result{
				Error: &errs.PDFError{
					Err:  errors.New("file too large"),
					Code: errs.ErrTooLarge,
				},
			}
		}

		return Result{
			Error: &errs.PDFError{
				Err:  errors.New("invalid request body"),
				Code: errs.ErrBadInput,
			},
		}
	}

	payload := append(header, body...)
	res.Metadata.Reader = bytes.NewReader(payload)
	res.Metadata.Size = int64(len(payload))

	return res
}
