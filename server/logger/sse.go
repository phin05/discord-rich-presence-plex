package logger

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const subMaxEntries = 100

var (
	subs   []chan entry
	subsMu sync.RWMutex
)

func subscribe() <-chan entry {
	subsMu.Lock()
	defer subsMu.Unlock()
	sub := make(chan entry, subMaxEntries)
	subs = append(subs, sub)
	return sub
}

func unsubscribe(ch <-chan entry) {
	subsMu.Lock()
	defer subsMu.Unlock()
	for i, sub := range subs {
		if sub == ch {
			close(sub)
			subs = append(subs[:i], subs[i+1:]...)
			return
		}
	}
}

func notifySubs(entry entry) {
	subsMu.RLock()
	defer subsMu.RUnlock()
	for _, sub := range subs {
		// Never block on slow/disconnected subscribers
		select {
		case sub <- entry:
		default:
		}
	}
}

func SseHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		Error(nil, "Streaming unsupported")
		return
	}
	_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	sub := subscribe()
	defer unsubscribe(sub)
	entries := buffer.getAll()
	var lastInitSentId uint64
	for _, entry := range entries {
		if entry.Id > lastInitSentId {
			lastInitSentId = entry.Id
		}
		if r.Context().Err() != nil {
			return
		}
		if err := sendSse(w, entry); err != nil {
			Error(err, "Error sending log entry")
			return
		}
		flusher.Flush()
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case entry, open := <-sub:
			if !open {
				return
			}
			if entry.Id <= lastInitSentId {
				continue
			}
			if err := sendSse(w, entry); err != nil {
				Error(err, "Error sending log entry")
				return
			}
			flusher.Flush()
		}
	}
}

func sendSse(w http.ResponseWriter, entry entry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}
