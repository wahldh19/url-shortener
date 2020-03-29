package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreate(t *testing.T) {
	t.Skip("TODO")
}

func TestRedirect(t *testing.T) {
	t.Skip("TODO")
}

func TestStats(t *testing.T) {
	t.Skip("TODO")
}

func TestStatus(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(status)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf(
			"expected status code %d but got %d",
			http.StatusOK, status,
		)
	}
}
