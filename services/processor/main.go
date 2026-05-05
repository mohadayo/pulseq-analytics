package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Event struct {
	ID        string                 `json:"id"`
	EventType string                 `json:"event_type"`
	Source    string                 `json:"source"`
	Data     map[string]interface{} `json:"data"`
	Received string                 `json:"received_at"`
}

type ProcessedEvent struct {
	Event
	ProcessedAt string   `json:"processed_at"`
	Tags        []string `json:"tags"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

var (
	processedEvents []ProcessedEvent
	mu              sync.RWMutex
	logger          *log.Logger
)

func init() {
	logger = log.New(os.Stdout, "processor ", log.LstdFlags|log.Lmsgprefix)
}

func ProcessEvent(event Event) ProcessedEvent {
	tags := classifyEvent(event)
	return ProcessedEvent{
		Event:       event,
		ProcessedAt: time.Now().UTC().Format(time.RFC3339),
		Tags:        tags,
	}
}

func classifyEvent(event Event) []string {
	var tags []string
	switch strings.ToLower(event.EventType) {
	case "click":
		tags = append(tags, "interaction", "ui")
	case "purchase":
		tags = append(tags, "transaction", "revenue")
	case "signup":
		tags = append(tags, "acquisition", "user")
	case "error":
		tags = append(tags, "incident", "alert")
	default:
		tags = append(tags, "general")
	}
	if event.Source != "" {
		tags = append(tags, "source:"+event.Source)
	}
	return tags
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{
		Status:    "healthy",
		Service:   "processor",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		logger.Printf("Failed to decode event: %v", err)
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if event.ID == "" || event.EventType == "" {
		http.Error(w, `{"error":"id and event_type are required"}`, http.StatusBadRequest)
		return
	}
	processed := ProcessEvent(event)
	mu.Lock()
	processedEvents = append(processedEvents, processed)
	mu.Unlock()
	logger.Printf("Processed event: id=%s type=%s tags=%v", processed.ID, processed.EventType, processed.Tags)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(processed)
}

func listProcessedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	mu.RLock()
	defer mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	if processedEvents == nil {
		json.NewEncoder(w).Encode([]ProcessedEvent{})
		return
	}
	json.NewEncoder(w).Encode(processedEvents)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/process", processHandler)
	mux.HandleFunc("/processed", listProcessedHandler)
	logger.Printf("Starting processor on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Fatalf("Server failed: %v", err)
	}
}
