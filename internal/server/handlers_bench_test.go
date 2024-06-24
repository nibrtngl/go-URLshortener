package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkShortenURLHandler(b *testing.B) {
	server := setupServerForTesting()
	requestBody := `https://example.com`
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(requestBody))
		b.StartTimer()

		_, _ = server.App.Test(req, -1)
	}
}

func BenchmarkRedirectToOriginalURL(b *testing.B) {
	server := setupServerForTesting()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodGet, "/1", nil)
		b.StartTimer()

		_, _ = server.App.Test(req, -1)
	}
}

func BenchmarkShortenAPIHandler(b *testing.B) {
	server := setupServerForTesting()
	requestBody := `{"url": "https://example.com"}`
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(requestBody))
		b.StartTimer()

		_, _ = server.App.Test(req, -1)
	}
}

func BenchmarkGetUserURLsHandler(b *testing.B) {
	server := setupServerForTesting()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		b.StartTimer()

		_, _ = server.App.Test(req, -1)
	}
}

func BenchmarkDeleteURLsHandler(b *testing.B) {
	server := setupServerForTesting()
	requestBody := `["url1", "url2"]`
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBufferString(requestBody))
		b.StartTimer()

		_, _ = server.App.Test(req, -1)
	}
}

func BenchmarkShortenBatchURLHandler(b *testing.B) {
	server := setupServerForTesting()
	requestBody := `[{"correlation_id": "1", "original_url": "https://example.com"}]`
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(requestBody))
		b.StartTimer()

		_, _ = server.App.Test(req, -1)
	}
}

func BenchmarkPingHandler(b *testing.B) {
	server := setupServerForTesting()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		b.StartTimer()

		_, _ = server.App.Test(req, -1)
	}
}
