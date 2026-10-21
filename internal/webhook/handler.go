// Package webhook receives payment provider webhooks.
package webhook

import (
	"net/http"

	"github.com/ainsleyclark/hookrelay/internal/store"
)

// Handler receives webhook events and saves them to a Store.
type Handler struct {
	store store.Store
}

// NewHandler returns a Handler that saves events to s.
func NewHandler(s store.Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
