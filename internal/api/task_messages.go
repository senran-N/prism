package api

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// GET /api/tasks/{id}/messages — fetch agent conversation from SC
func (s *Server) handleTaskMessages(w http.ResponseWriter, r *http.Request) {
	ticketID := r.PathValue("id")

	acct := s.pool.GetTicketAccount(ticketID)
	if acct == nil || acct.Client == nil {
		writeError(w, 404, "task not found")
		return
	}

	impl, err := acct.Client.GetImplementation(ticketID)
	if err != nil || impl == nil {
		writeError(w, 404, "no messages yet: "+err.Error())
		return
	}

	writeJSON(w, map[string]any{
		"ticket_id": ticketID,
		"status":    impl.Status,
		"messages":  impl.Messages,
	})
}

// POST /api/tasks/{id}/message — send follow-up to agent
func (s *Server) handleTaskSendMessage(w http.ResponseWriter, r *http.Request) {
	ticketID := r.PathValue("id")

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Content) == "" {
		writeError(w, 400, "content is required")
		return
	}

	// Find the account + conversation for this ticket
	for _, acct := range s.pool.ListAll() {
		if acct.Client == nil {
			continue
		}

		// Get the implementation page to find conversation ID
		html, err := acct.Client.GetTicketHTML(ticketID)
		if err != nil || html == "" {
			continue
		}

		// Extract conversation ID from the page
		convRe := regexp.MustCompile(`/conversations/([A-Za-z0-9]+)/messages`)
		m := convRe.FindStringSubmatch(html)
		if m == nil {
			continue
		}
		convID := m[1]

		err = acct.Client.SendMessage(convID, req.Content)
		if err != nil {
			log.Printf("[task] send message error: %v", err)
			writeError(w, 500, "failed to send message")
			return
		}

		log.Printf("[task] message sent to ticket %s conv %s", ticketID, convID)
		writeJSON(w, map[string]bool{"sent": true})
		return
	}

	writeError(w, 404, "task not found")
}
