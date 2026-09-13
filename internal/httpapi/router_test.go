package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterRequiresBearerToken(t *testing.T) {
	router := New(nil, "secret", "https://ui.example")
	request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestRouterAnswersCORSPreflightWithoutAuthentication(t *testing.T) {
	router := New(nil, "secret", "https://ui.example")
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/equipments", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", response.Code)
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "https://ui.example" {
		t.Fatalf("unexpected CORS origin: %q", response.Header().Get("Access-Control-Allow-Origin"))
	}
}
