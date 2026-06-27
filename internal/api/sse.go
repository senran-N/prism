package api

import (
	"fmt"
	"net/http"
	"time"
)

// GET /api/events — SSE stream for real-time task updates
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", 500)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Send initial pool stats
	stats := s.pool.Stats()
	fmt.Fprintf(w, "event: pool\ndata: {\"ready\":%d,\"active\":%d,\"exhausted\":%d,\"total_credits\":%.2f}\n\n",
		stats.Ready, stats.Active, stats.Exhausted, stats.TotalCredits)
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			stats := s.pool.Stats()
			fmt.Fprintf(w, "event: pool\ndata: {\"ready\":%d,\"active\":%d,\"exhausted\":%d,\"total_credits\":%.2f}\n\n",
				stats.Ready, stats.Active, stats.Exhausted, stats.TotalCredits)
			flusher.Flush()
		}
	}
}
