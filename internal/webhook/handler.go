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

	_, err := h.store.Get(r.Context(), ev.ID)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	if !errors.Is(err, store.ErrNotFound) {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := h.store.Save(r.Context(), ev); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
