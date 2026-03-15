package load

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpx "Alice088/pdf-summarize/internal/http"
	"Alice088/pdf-summarize/pkg/env"

	"github.com/go-chi/chi/v5"
)

func newTestRouter(handler http.HandlerFunc) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	cfg := env.Config{
		MaxUser:              5,
		Timeout:              5 * time.Second,
		Origins:              []string{"http://localhost"},
		AllowContentEncoding: []string{"gzip", "identity"},
	}

	r := chi.NewRouter()

	httpx.UpMiddlewares(r, cfg, logger)

	r.Post("/load", handler)

	return r
}

func TestLoadHandler(t *testing.T) {
	router := newTestRouter(Load(nil))

	tests := []struct {
		name       string
		body       []byte
		statusCode int
	}{
		{
			name:       "valid pdf",
			body:       append([]byte("%PDF-"), bytes.Repeat([]byte("a"), 100)...),
			statusCode: http.StatusOK,
		},
		{
			name:       "invalid signature",
			body:       []byte("NOTPDF"),
			statusCode: http.StatusUnsupportedMediaType,
		},
		{
			name:       "empty body",
			body:       []byte{},
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "small corrupted pdf",
			body:       []byte("%PDF-"),
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest(
				http.MethodPost,
				"/load",
				bytes.NewReader(tt.body),
			)

			req.Header.Set("Content-Type", "application/pdf")

			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.statusCode {
				t.Fatalf("expected %d got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestLoadHandler_LargeFile(t *testing.T) {
	router := newTestRouter(Load(nil))

	large := append(
		[]byte("%PDF-"),
		bytes.Repeat([]byte("A"), 6<<20)...,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/load",
		bytes.NewReader(large),
	)

	req.Header.Set("Content-Type", "application/pdf")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("large file should not be accepted")
	}
}

func TestLoadHandler_RandomData(t *testing.T) {
	router := newTestRouter(Load(nil))

	data := bytes.Repeat([]byte("X"), 1000)

	req := httptest.NewRequest(
		http.MethodPost,
		"/load",
		bytes.NewReader(data),
	)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code == http.StatusInternalServerError {
		t.Fatalf("handler crashed on random input")
	}
}

func TestLoad_Throttle(t *testing.T) {
	router := newTestRouter(Load(nil))

	body := append([]byte("%PDF-"), bytes.Repeat([]byte("A"), 100)...)

	done := make(chan bool)

	for i := 0; i < 20; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodPost, "/load", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/pdf")

			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			done <- true
		}()
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestLoad_CORS(t *testing.T) {
	router := newTestRouter(Load(nil))

	req := httptest.NewRequest(http.MethodOptions, "/load", nil)
	req.Header.Set("Origin", "http://localhost")
	req.Header.Set("Access-Control-Request-Method", "POST")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("cors preflight failed")
	}
}

func BenchmarkLoadHandler(b *testing.B) {
	router := newTestRouter(Load(nil))

	body := append(
		[]byte("%PDF-"),
		bytes.Repeat([]byte("A"), 1000)...,
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		req := httptest.NewRequest(
			http.MethodPost,
			"/load",
			bytes.NewReader(body),
		)

		req.Header.Set("Content-Type", "application/pdf")

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
	}
}
