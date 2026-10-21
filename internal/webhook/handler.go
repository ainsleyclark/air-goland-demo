//go:build !fixed

package webhook

import (
	"encoding/json"
	"errors"
	"log/slog"
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

// ServeHTTP decodes the event in the request body and saves it, replying
// 202 Accepted for a new event and 200 OK for a redelivery.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var ev store.Event

	// ISSUE: 	The body is decoded with no size limit, so a client can send an arbitrarily large payload.
	// DEMO: 	Code Review Skill
	// FIX: 	Add const maxBodyBytes = 1 << 20 and decode from http.MaxBytesReader(w, r.Body, maxBodyBytes). See handler_fixed.go.
	if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
		slog.WarnContext(ctx, "decoding webhook", "error", err)
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := validateEventID(ev.ID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ev.Provider = r.PathValue("provider")

	_, err := h.store.Get(ctx, ev.ID)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	if !errors.Is(err, store.ErrNotFound) {
		slog.ErrorContext(ctx, "loading webhook", "id", ev.ID, "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// ISSUE: 	Two deliveries of one event can both pass the Get check, so the second Save gets a 500 and the provider retries forever.
	// DEMO: 	Air diff feedback on Agent 1: "return 200 for duplicates, or the provider retries forever."
	// FIX: 	Return 200 when errors.Is(err, store.ErrDuplicate), and drop the Get check. See handler_fixed.go.
	if err = h.store.Save(ctx, ev); err != nil {
		slog.ErrorContext(ctx, "saving webhook", "id", ev.ID, "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
