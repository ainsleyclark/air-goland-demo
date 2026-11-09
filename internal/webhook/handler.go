package webhook

import (
	"encoding/json"
	"net/http"
	"regexp"

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

// ServeHTTP decodes the event in the request body and saves it.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var ev store.Event
	if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if !validEventID(ev.ID) {
		http.Error(w, "invalid event ID", http.StatusBadRequest)
		return
	}

	if err := h.store.Save(r.Context(), ev); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// validEventID reports whether id looks like a provider event ID.
func validEventID(id string) bool {
	re := regexp.MustCompile(`^evt_[A-Za-z0-9]{16,64}$`)
	return re.MatchString(id)
}
