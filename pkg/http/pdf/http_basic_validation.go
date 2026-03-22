package pdf

import (
	"Alice088/pdf-summarize/pkg/size"
	"bytes"
	"errors"
	"io"
	"net/http"
)

const pdfHeaderSize = 5

type Result struct {
	Valid    Valid
	Metadata SafeMetadata
}

type Valid struct {
	Error error
	Code  *int
}

type SafeMetadata struct {
	Size   int64
	Reader io.Reader
}

func BasicValid(r *http.Request) Result {
	res := Result{
		Valid: Valid{},
	}

	if r.ContentLength > size.MB5 {
		return Result{
			Valid: Valid{
				Error: errors.New("file too large"),
				Code:  new(http.StatusBadRequest),
			},
		}
	}

	header := make([]byte, pdfHeaderSize)
	if _, err := io.ReadFull(r.Body, header); err != nil {
		return Result{
			Valid: Valid{
				Error: errors.New("invalid request body"),
				Code:  new(http.StatusBadRequest),
			},
		}
	}

	if string(header) != "%PDF-" {
		return Result{
			Valid: Valid{
				Error: errors.New("invalid pdf file"),
				Code:  new(http.StatusUnsupportedMediaType),
			},
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return Result{
				Valid: Valid{
					Error: errors.New("file too large"),
					Code:  new(http.StatusBadRequest),
				},
			}
		}

		return Result{
			Valid: Valid{
				Error: errors.New("invalid request body"),
				Code:  new(http.StatusBadRequest),
			},
		}
	}

	payload := append(header, body...)
	res.Metadata.Reader = bytes.NewReader(payload)
	res.Metadata.Size = int64(len(payload))

	return res
}
