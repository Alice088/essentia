package pdf

import (
	"Alice088/pdf-summarize/pkg/size"
	"bytes"
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

	res := BasicValid(req)

	if res.Valid.Error != nil {
		t.Fatalf("expected no error, got %v", res.Valid.Error)
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
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("%PDF-")))
	req.ContentLength = size.MB5 + 1

	res := BasicValid(req)

	if res.Valid.Error == nil {
		t.Fatal("expected error")
	}

	if *res.Valid.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", *res.Valid.Code)
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

	res := BasicValid(req)

	if res.Valid.Error == nil {
		t.Fatal("expected error")
	}

	if *res.Valid.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", *res.Valid.Code)
	}

	if res.Valid.Error.Error() != "file too large" {
		t.Fatalf("expected file too large error, got %q", res.Valid.Error.Error())
	}
}

// TestBasicValid_InvalidHeader verifies that non-PDF payloads are rejected
// before the service layer receives them.
func TestBasicValid_InvalidHeader(t *testing.T) {
	body := []byte("NOTPDF")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.ContentLength = int64(len(body))

	res := BasicValid(req)

	if res.Valid.Error == nil {
		t.Fatal("expected error")
	}

	if *res.Valid.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", *res.Valid.Code)
	}
}

// TestBasicValid_ShortBody verifies that payloads shorter than the PDF header
// are treated as malformed request bodies.
func TestBasicValid_ShortBody(t *testing.T) {
	body := []byte("123")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.ContentLength = int64(len(body))

	res := BasicValid(req)

	if res.Valid.Error == nil {
		t.Fatal("expected error")
	}

	if *res.Valid.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", *res.Valid.Code)
	}
}

// TestBasicValid_UnknownSizeUsesActualBufferedLength verifies that when
// Content-Length is unknown, validation buffers the safe payload and reports
// the actual size that will be passed downstream.
func TestBasicValid_UnknownSizeUsesActualBufferedLength(t *testing.T) {
	body := []byte("%PDF-test")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.ContentLength = -1

	res := BasicValid(req)

	if res.Valid.Error != nil {
		t.Fatalf("unexpected error: %v", res.Valid.Error)
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
