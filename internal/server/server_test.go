package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusCreated, map[string]string{"key": "value"})

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["key"] != "value" {
		t.Errorf("body[key] = %q, want %q", body["key"], "value")
	}
}

func TestParseLimit(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", defaultPageLimit},
		{"10", 10},
		{"1", 1},
		{"100", 100},
		{"200", maxPageLimit},
		{"0", defaultPageLimit},
		{"-5", defaultPageLimit},
		{"abc", defaultPageLimit},
	}

	for _, tt := range tests {
		got := parseLimit(tt.input)
		if got != tt.want {
			t.Errorf("parseLimit(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
