package errors

import (
	"net/http"
)

type PDFErrorCode string

const (
	ErrTooLarge PDFErrorCode = "too_large"
	ErrNotPDF   PDFErrorCode = "not_pdf"
	ErrBadInput PDFErrorCode = "bad_input"
)

type PDFError struct {
	Err  error
	Code PDFErrorCode
}

func (e PDFError) Error() string {
	return e.Err.Error()
}

func (e PDFError) StatusCode() int {
	switch e.Code {
	case ErrTooLarge:
		return http.StatusBadRequest
	case ErrNotPDF:
		return http.StatusUnsupportedMediaType
	case ErrBadInput:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (e PDFError) SafeMessage() string {
	switch e.Code {
	case ErrTooLarge:
		return "file too large"
	case ErrNotPDF:
		return "file not pdf"
	case ErrBadInput:
		return "invalid file"
	default:
		return "internal server error"
	}
}
