package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create a new request: %v", err)
	}
	req.Header.Set("Origin", "example.com")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	corsHandler := CORS()(handler)
	corsHandler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, status)
	}
	if cors := rr.Header().Get("Access-Control-Allow-Origin"); cors != "example.com" {
		t.Errorf("Expected header Access-Control-Allow-Origin: %s, but got %s", "example.com", cors)
	}
}
