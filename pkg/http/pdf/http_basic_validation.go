package pdf

import (
	"Alice088/pdf-summarize/pkg/size"
	"bytes"
	"errors"
	"io"
	"net/http"
)

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
	header := make([]byte, 5)
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

	res.Metadata.Reader = io.MultiReader(
		bytes.NewReader(header),
		r.Body,
	)

	res.Metadata.Size = r.ContentLength
	if res.Metadata.Size <= 0 {
		res.Metadata.Size = -1
	}

	return res
}
