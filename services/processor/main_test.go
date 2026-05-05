package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func resetStore() {
	mu.Lock()
	processedEvents = nil
	mu.Unlock()
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	healthHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp HealthResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Status != "healthy" {
		t.Fatalf("expected healthy, got %s", resp.Status)
	}
	if resp.Service != "processor" {
		t.Fatalf("expected processor, got %s", resp.Service)
	}
}

func TestProcessHandler(t *testing.T) {
	resetStore()
	event := Event{ID: "e1", EventType: "click", Source: "web", Data: map[string]interface{}{"page": "/home"}}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/process", bytes.NewReader(body))
	w := httptest.NewRecorder()
	processHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var processed ProcessedEvent
	json.NewDecoder(w.Body).Decode(&processed)
	if processed.ID != "e1" {
		t.Fatalf("expected id e1, got %s", processed.ID)
	}
	if len(processed.Tags) == 0 {
		t.Fatal("expected tags to be populated")
	}
}

func TestProcessHandlerInvalidBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/process", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()
	processHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProcessHandlerMissingFields(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"source": "web"})
	req := httptest.NewRequest(http.MethodPost, "/process", bytes.NewReader(body))
	w := httptest.NewRecorder()
	processHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProcessHandlerMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/process", nil)
	w := httptest.NewRecorder()
	processHandler(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestListProcessedHandler(t *testing.T) {
	resetStore()
	event := Event{ID: "e2", EventType: "purchase", Source: "mobile"}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/process", bytes.NewReader(body))
	w := httptest.NewRecorder()
	processHandler(w, req)

	req = httptest.NewRequest(http.MethodGet, "/processed", nil)
	w = httptest.NewRecorder()
	listProcessedHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var events []ProcessedEvent
	json.NewDecoder(w.Body).Decode(&events)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestListProcessedEmpty(t *testing.T) {
	resetStore()
	req := httptest.NewRequest(http.MethodGet, "/processed", nil)
	w := httptest.NewRecorder()
	listProcessedHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var events []ProcessedEvent
	json.NewDecoder(w.Body).Decode(&events)
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

func TestClassifyEvent(t *testing.T) {
	tests := []struct {
		eventType string
		expected  string
	}{
		{"click", "interaction"},
		{"purchase", "transaction"},
		{"signup", "acquisition"},
		{"error", "incident"},
		{"unknown", "general"},
	}
	for _, tc := range tests {
		tags := classifyEvent(Event{EventType: tc.eventType, Source: "test"})
		found := false
		for _, tag := range tags {
			if tag == tc.expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("event type %s: expected tag %s in %v", tc.eventType, tc.expected, tags)
		}
	}
}
