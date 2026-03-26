package validation

import (
	errs "Alice088/essentia/pkg/errors"
	"Alice088/essentia/pkg/size"
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestBasicValid_Success verifies that a valid PDF payload passes validation,
// preserves the original bytes, and reports the exact payload size.
func TestBasicValid_Success(t *testing.T) {
	body := []byte("%PDF-hello world")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.ContentLength = int64(len(body))

	res := PDF(PDFInput{
		Size:   req.ContentLength,
		Reader: bytes.NewBuffer(body),
	})

	if res.Error != nil {
		t.Fatalf("expected no error, got %v", res.Error)
	}

	if res.Metadata.Size != int64(len(body)) {
		t.Fatalf("expected size %d, got %d", len(body), res.Metadata.Size)
	}

	out, err := io.ReadAll(res.Metadata.Reader)
	if err != nil {
		t.Fatalf("failed to read reader: %v", err)
	}

	if !bytes.Equal(out, body) {
		t.Fatalf("reader content mismatch")
	}
}

// TestBasicValid_FileTooLargeContentLength verifies that requests declaring
// a Content-Length larger than the allowed limit are rejected immediately.
func TestBasicValid_FileTooLargeContentLength(t *testing.T) {
	body := []byte("%PDF-hello world")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.ContentLength = size.MB5 + 1

	res := PDF(PDFInput{
		Size:   req.ContentLength,
		Reader: bytes.NewBuffer(body),
	})

	if res.Error == nil {
		t.Fatal("expected error")
	}

	if err, ok := errors.AsType[*errs.PDFError](res.Error); !ok {
		t.Fatalf("expected PDFError, got %v", res.Error)
	} else {
		if err.StatusCode() != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", err.StatusCode())
		}
	}

}

// TestBasicValid_FileTooLargeUnknownSize verifies that oversized streamed
// requests are rejected when the middleware-backed MaxBytesReader trips.
func TestBasicValid_FileTooLargeUnknownSize(t *testing.T) {
	body := append([]byte("%PDF-"), bytes.Repeat([]byte("a"), size.MB5)...)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.ContentLength = -1
	rr := httptest.NewRecorder()
	req.Body = http.MaxBytesReader(rr, req.Body, size.MB5)

	res := PDF(PDFInput{
		Size:   req.ContentLength,
		Reader: req.Body,
	})

	if res.Error == nil {
		t.Fatal("expected error")
	}

	if err, ok := errors.AsType[*errs.PDFError](res.Error); !ok {
		t.Fatalf("expected PDFError, got %v", res.Error)
	} else {
		if err.StatusCode() != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", err.StatusCode())
		}

		if err.Code != errs.ErrTooLarge {
			t.Fatalf("expected file too large error, got %q", err.Code)
		}
	}
}

// TestBasicValid_InvalidHeader verifies that non-PDF payloads are rejected
// before the service layer receives them.
func TestBasicValid_InvalidHeader(t *testing.T) {
	body := []byte("NOTPDF")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.ContentLength = int64(len(body))

	res := PDF(PDFInput{
		Size:   req.ContentLength,
		Reader: bytes.NewBuffer(body),
	})

	if res.Error == nil {
		t.Fatal("expected error")
	}

	if err, ok := errors.AsType[*errs.PDFError](res.Error); !ok {
		t.Fatalf("expected PDFError, got %v", res.Error)
	} else {
		if err.StatusCode() != http.StatusUnsupportedMediaType {
			t.Fatalf("expected 415, got %d", err.StatusCode())
		}
	}
}

// TestBasicValid_ShortBody verifies that payloads shorter than the PDF header
// are treated as malformed request bodies.
func TestBasicValid_ShortBody(t *testing.T) {
	body := []byte("123")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.ContentLength = int64(len(body))

	res := PDF(PDFInput{
		Size:   req.ContentLength,
		Reader: bytes.NewBuffer(body),
	})

	if res.Error == nil {
		t.Fatal("expected error")
	}

	if err, ok := errors.AsType[*errs.PDFError](res.Error); !ok {
		t.Fatalf("expected PDFError, got %v", res.Error)
	} else {
		if err.StatusCode() != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", err.StatusCode())
		}
	}
}

// TestBasicValid_UnknownSizeUsesActualBufferedLength verifies that when
// Content-Length is unknown, validation buffers the safe payload and reports
// the actual size that will be passed downstream.
func TestBasicValid_UnknownSizeUsesActualBufferedLength(t *testing.T) {
	body := []byte("%PDF-test")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.ContentLength = -1

	res := PDF(PDFInput{
		Size:   req.ContentLength,
		Reader: bytes.NewBuffer(body),
	})

	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}

	if res.Metadata.Size != int64(len(body)) {
		t.Fatalf("expected size %d, got %d", len(body), res.Metadata.Size)
	}

	out, err := io.ReadAll(res.Metadata.Reader)
	if err != nil {
		t.Fatalf("failed to read reader: %v", err)
	}

	if !bytes.Equal(out, body) {
		t.Fatalf("reader content mismatch")
	}
}
