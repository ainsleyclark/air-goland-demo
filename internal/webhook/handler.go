// Package webhook receives payment provider webhooks.
package webhook

import (
	"encoding/json"
	"errors"
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
	var ev store.Event
	if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateEventID(ev.ID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ev.Provider = r.PathValue("provider")

	err := h.store.Save(r.Context(), ev)
	switch {
	case errors.Is(err, store.ErrDuplicate):
		w.WriteHeader(http.StatusOK)
	case err != nil:
		http.Error(w, "internal error", http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusAccepted)
	}
}
