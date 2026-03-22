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

	"github.com/google/uuid"
)

type pdfServiceStub struct {
	createJob func(ctx context.Context, r io.Reader, size int64) (uuid.UUID, error)
}

func (s pdfServiceStub) CreateJob(ctx context.Context, r io.Reader, size int64) (uuid.UUID, error) {
	return s.createJob(ctx, r, size)
}

func TestHandlerLoad_Success(t *testing.T) {
	body := []byte("%PDF-hello world")
	expectedJobID := uuid.New()
	called := false

	h := &Handler{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		PDFService: pdfServiceStub{createJob: func(ctx context.Context, r io.Reader, size int64) (uuid.UUID, error) {
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
		}},
	}

	req := httptest.NewRequest(http.MethodPost, "/pdf/load", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/pdf")
	req.ContentLength = int64(len(body))
	rr := httptest.NewRecorder()

	h.Load().ServeHTTP(rr, req)

	if !called {
		t.Fatal("expected CreateJob to be called")
	}

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var got map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got["job_id"] != expectedJobID.String() {
		t.Fatalf("expected job_id %q, got %q", expectedJobID.String(), got["job_id"])
	}
}

func TestHandlerLoad_InvalidPDF(t *testing.T) {
	called := false
	body := []byte("NOTPDF")

	h := &Handler{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		PDFService: pdfServiceStub{createJob: func(ctx context.Context, r io.Reader, size int64) (uuid.UUID, error) {
			called = true
			return uuid.Nil, nil
		}},
	}

	req := httptest.NewRequest(http.MethodPost, "/pdf/load", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/pdf")
	req.ContentLength = int64(len(body))
	rr := httptest.NewRecorder()

	h.Load().ServeHTTP(rr, req)

	if called {
		t.Fatal("expected CreateJob not to be called")
	}

	if rr.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status %d, got %d", http.StatusUnsupportedMediaType, rr.Code)
	}

	var got map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got["error"] != "invalid pdf file" {
		t.Fatalf("expected invalid pdf error, got %q", got["error"])
	}
}

func TestHandlerLoad_CreateJobError(t *testing.T) {
	body := []byte("%PDF-hello world")
	called := false

	h := &Handler{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		PDFService: pdfServiceStub{createJob: func(ctx context.Context, r io.Reader, size int64) (uuid.UUID, error) {
			called = true

			data, err := io.ReadAll(r)
			if err != nil {
				t.Fatalf("failed to read request body: %v", err)
			}
			if !bytes.Equal(data, body) {
				t.Fatalf("expected body %q, got %q", body, data)
			}

			return uuid.Nil, errors.New("boom")
		}},
	}

	req := httptest.NewRequest(http.MethodPost, "/pdf/load", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/pdf")
	req.ContentLength = int64(len(body))
	rr := httptest.NewRecorder()

	h.Load().ServeHTTP(rr, req)

	if !called {
		t.Fatal("expected CreateJob to be called")
	}

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	var got map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got["error"] != "failed to process file" {
		t.Fatalf("expected process file error, got %q", got["error"])
	}
}
