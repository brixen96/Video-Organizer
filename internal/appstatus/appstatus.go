package appstatus

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"video-organizer/internal/database"
)

// AppEvent represents a status event emitted by the application.
// Standard categories should be prefixed with brackets:
// - [Task] - Long-running operations like scanning, initialization
// - [Progress] - Step-by-step updates within tasks
// - [Info] - General information
// - [Warning] - Non-fatal issues
// - [Error] - Failures
type AppEvent struct {
	Type      string `json:"type"`      // short type identifier
	Category  string `json:"category"`  // e.g., [Task], [Progress], [Info], [Warning], [Error]
	Message   string `json:"message"`   // human-readable message
	Level     string `json:"level"`     // info, warning, error
	Timestamp int64  `json:"timestamp"` // unix ms
}

var (
	mu             sync.Mutex
	subscribers    = map[int]chan AppEvent{}
	nextSubscriber = 1
)

// Emit broadcasts an event to all connected subscribers. Non-blocking.
func Emit(evt AppEvent) {
	evt.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)

	// Persist to DB (best-effort)
	go func(e AppEvent) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered while persisting event: %v", r)
			}
		}()
		db := database.GetDB()
		if db != nil {
			_, err := db.Exec("INSERT INTO monitor_events(type, category, message, level, timestamp) VALUES(?, ?, ?, ?, ?)", e.Type, e.Category, e.Message, e.Level, e.Timestamp)
			if err != nil {
				// non-fatal
				log.Printf("Failed to persist monitor event: %v", err)
			}
		}
	}(evt)

	mu.Lock()
	defer mu.Unlock()
	for id, ch := range subscribers {
		select {
		case ch <- evt:
		default:
			// subscriber is slow; drop the event rather than block
			log.Printf("Dropping event for subscriber %d: %+v", id, evt)
		}
	}
}

// EmitInfo emits an informational event
func EmitInfo(category, msg string) {
	Emit(AppEvent{Type: "info", Category: category, Message: msg, Level: "info"})
}

// EmitWarning emits a warning event
func EmitWarning(category, msg string) {
	Emit(AppEvent{Type: "warning", Category: category, Message: msg, Level: "warning"})
}

// EmitError emits an error event
func EmitError(category, msg string) {
	Emit(AppEvent{Type: "error", Category: category, Message: msg, Level: "error"})
}

// SSEHandler upgrades the connection to Server-Sent Events and streams events
func SSEHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Set headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create subscriber channel
	ch := make(chan AppEvent, 16)

	mu.Lock()
	id := nextSubscriber
	nextSubscriber++
	subscribers[id] = ch
	mu.Unlock()

	// Remove subscriber on exit
	notify := r.Context().Done()

	// Send a warmup event
	warm := AppEvent{Type: "connect", Category: "monitor", Message: "connected", Level: "info", Timestamp: time.Now().UnixNano() / int64(time.Millisecond)}
	b, _ := json.Marshal(warm)
	fmt.Fprintf(w, "data: %s\n\n", b)
	flusher.Flush()

	for {
		select {
		case <-notify:
			mu.Lock()
			delete(subscribers, id)
			mu.Unlock()
			close(ch)
			return
		case evt := <-ch:
			// Marshal and send
			b, err := json.Marshal(evt)
			if err != nil {
				log.Println("Failed to marshal event:", err)
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", b)
			flusher.Flush()
		}
	}
}
