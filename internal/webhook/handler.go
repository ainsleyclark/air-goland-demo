// Package webhook receives payment provider webhooks.
package webhook

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ainsleyclark/hookrelay/internal/store"
)

const maxBodyBytes = 1 << 20

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
	body := http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(body).Decode(&ev); err != nil {
		slog.WarnContext(r.Context(), "decoding webhook", "error", err)
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
		}
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
		slog.ErrorContext(r.Context(), "saving webhook", "id", ev.ID, "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusAccepted)
	}
}
