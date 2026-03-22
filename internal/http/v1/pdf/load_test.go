package pdf

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"Alice088/pdf-summarize/pkg/size"
	"github.com/google/uuid"
)

type pdfServiceStub struct {
	createJob func(ctx context.Context, r io.Reader, size int64) (uuid.UUID, error)
}

func (s pdfServiceStub) CreateJob(ctx context.Context, r io.Reader, size int64) (uuid.UUID, error) {
	return s.createJob(ctx, r, size)
}

func newTestHandler(t *testing.T, createJob func(ctx context.Context, r io.Reader, size int64) (uuid.UUID, error)) *Handler {
	t.Helper()

	return &Handler{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		PDFService: pdfServiceStub{createJob: func(ctx context.Context, r io.Reader, size int64) (uuid.UUID, error) {
			if createJob == nil {
				t.Fatal("unexpected CreateJob call")
			}

			return createJob(ctx, r, size)
		}},
	}
}

func decodeResponseBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]string {
	t.Helper()

	var got map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return got
}

func assertJSONResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int) map[string]string {
	t.Helper()

	if rr.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, rr.Code)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	return decodeResponseBody(t, rr)
}

func TestHandlerLoad_Success(t *testing.T) {
	body := []byte("%PDF-hello world")
	expectedJobID := uuid.New()
	called := false

	h := newTestHandler(t, func(ctx context.Context, r io.Reader, size int64) (uuid.UUID, error) {
		called = true

		if size != int64(len(body)) {
			t.Fatalf("expected size %d, got %d", len(body), size)
		}

		data, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if !bytes.Equal(data, body) {
			t.Fatalf("expected body %q, got %q", body, data)
		}

		return expectedJobID, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/pdf/load", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/pdf")
	req.ContentLength = int64(len(body))
	rr := httptest.NewRecorder()

	h.Load().ServeHTTP(rr, req)

	if !called {
		t.Fatal("expected CreateJob to be called")
	}

	got := assertJSONResponse(t, rr, http.StatusOK)
	if got["job_id"] != expectedJobID.String() {
		t.Fatalf("expected job_id %q, got %q", expectedJobID.String(), got["job_id"])
	}
}

func TestHandlerLoad_UnknownContentLengthUsesBufferedSize(t *testing.T) {
	body := []byte("%PDF-streamed body")
	called := false

	h := newTestHandler(t, func(ctx context.Context, r io.Reader, size int64) (uuid.UUID, error) {
		called = true

		if size != int64(len(body)) {
			t.Fatalf("expected size %d, got %d", len(body), size)
		}

		data, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if !bytes.Equal(data, body) {
			t.Fatalf("expected body %q, got %q", body, data)
		}

		return uuid.New(), nil
	})

	req := httptest.NewRequest(http.MethodPost, "/pdf/load", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/pdf")
	req.ContentLength = -1
	rr := httptest.NewRecorder()

	h.Load().ServeHTTP(rr, req)

	if !called {
		t.Fatal("expected CreateJob to be called")
	}

	_ = assertJSONResponse(t, rr, http.StatusOK)
}

func TestHandlerLoad_RejectsOversizedBodyWithUnknownContentLength(t *testing.T) {
	body := append([]byte("%PDF-"), bytes.Repeat([]byte("a"), size.MB5)...)
	h := newTestHandler(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/pdf/load", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/pdf")
	req.ContentLength = -1
	rr := httptest.NewRecorder()

	h.Load().ServeHTTP(rr, req)

	got := assertJSONResponse(t, rr, http.StatusBadRequest)
	if got["error"] != "file too large" {
		t.Fatalf("expected file too large error, got %q", got["error"])
	}
}

func TestHandlerLoad_InvalidPDF(t *testing.T) {
	body := []byte("NOTPDF")
	h := newTestHandler(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/pdf/load", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/pdf")
	req.ContentLength = int64(len(body))
	rr := httptest.NewRecorder()

	h.Load().ServeHTTP(rr, req)

	got := assertJSONResponse(t, rr, http.StatusUnsupportedMediaType)
	if got["error"] != "invalid pdf file" {
		t.Fatalf("expected invalid pdf error, got %q", got["error"])
	}
}

func TestHandlerLoad_ShortBody(t *testing.T) {
	body := []byte("123")
	h := newTestHandler(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/pdf/load", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/pdf")
	req.ContentLength = int64(len(body))
	rr := httptest.NewRecorder()

	h.Load().ServeHTTP(rr, req)

	got := assertJSONResponse(t, rr, http.StatusBadRequest)
	if got["error"] != "invalid request body" {
		t.Fatalf("expected invalid request body error, got %q", got["error"])
	}
}

func TestHandlerLoad_FileTooLarge(t *testing.T) {
	body := []byte("%PDF-too large")
	h := newTestHandler(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/pdf/load", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/pdf")
	req.ContentLength = size.MB5 + 1
	rr := httptest.NewRecorder()

	h.Load().ServeHTTP(rr, req)

	got := assertJSONResponse(t, rr, http.StatusBadRequest)
	if got["error"] != "file too large" {
		t.Fatalf("expected file too large error, got %q", got["error"])
	}
}

func TestHandlerLoad_CreateJobError(t *testing.T) {
	body := []byte("%PDF-hello world")
	called := false

	h := newTestHandler(t, func(ctx context.Context, r io.Reader, size int64) (uuid.UUID, error) {
		called = true

		data, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if !bytes.Equal(data, body) {
			t.Fatalf("expected body %q, got %q", body, data)
		}

		return uuid.Nil, errors.New("boom")
	})

	req := httptest.NewRequest(http.MethodPost, "/pdf/load", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/pdf")
	req.ContentLength = int64(len(body))
	rr := httptest.NewRecorder()

	h.Load().ServeHTTP(rr, req)

	if !called {
		t.Fatal("expected CreateJob to be called")
	}

	got := assertJSONResponse(t, rr, http.StatusInternalServerError)
	if got["error"] != "failed to process file" {
		t.Fatalf("expected process file error, got %q", got["error"])
	}
}
