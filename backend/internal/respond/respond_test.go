package respond_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"movie-tracker/internal/respond"
)

func TestJSON_SetsContentTypeAndStatus(t *testing.T) {
	w := httptest.NewRecorder()
	respond.JSON(w, http.StatusCreated, map[string]string{"key": "value"})

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json Content-Type, got %s", ct)
	}
}

func TestJSON_EncodesBody(t *testing.T) {
	w := httptest.NewRecorder()
	respond.JSON(w, http.StatusOK, map[string]int{"count": 42})

	var out map[string]int
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if out["count"] != 42 {
		t.Errorf("expected count=42, got %d", out["count"])
	}
}

func TestError_SetsErrorField(t *testing.T) {
	w := httptest.NewRecorder()
	respond.Error(w, http.StatusBadRequest, "bad input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	var out map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if out["error"] != "bad input" {
		t.Errorf("expected error='bad input', got %q", out["error"])
	}
}

func TestError_StatusCodes(t *testing.T) {
	cases := []struct {
		status int
		msg    string
	}{
		{http.StatusBadRequest, "missing field"},
		{http.StatusNotFound, "not found"},
		{http.StatusInternalServerError, "internal error"},
		{http.StatusConflict, "duplicate"},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		respond.Error(w, tc.status, tc.msg)
		if w.Code != tc.status {
			t.Errorf("status %d: expected %d, got %d", tc.status, tc.status, w.Code)
		}
	}
}
