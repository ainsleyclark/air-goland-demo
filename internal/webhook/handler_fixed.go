//go:build fixed

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

// ServeHTTP decodes the event in the request body and saves it, replying
// 202 Accepted for a new event and 200 OK for a redelivery.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var ev store.Event
	body := http.MaxBytesReader(w, r.Body, maxBodyBytes)

	if err := json.NewDecoder(body).Decode(&ev); err != nil {
		slog.WarnContext(ctx, "decoding webhook", "error", err)
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

	err := h.store.Save(ctx, ev)
	switch {
	case errors.Is(err, store.ErrDuplicate):
		w.WriteHeader(http.StatusOK)
	case err != nil:
		slog.ErrorContext(ctx, "saving webhook", "id", ev.ID, "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusAccepted)
	}
}
