//go:build fixed

package webhook

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ainsleyclark/hookrelay/internal/store"
)

// duplicateStore finds nothing on Get but reports every Save as a
// duplicate, as if another delivery of the same event won the race.
type duplicateStore struct{}

func (duplicateStore) Get(context.Context, string) (store.Event, error) {
	return store.Event{}, store.ErrNotFound
}

func (duplicateStore) Save(context.Context, store.Event) error {
	return store.ErrDuplicate
}

func TestHandler_ServeHTTP_Fixed(t *testing.T) {
	t.Parallel()

	t.Run("Racing duplicate acknowledged", func(t *testing.T) {
		t.Parallel()

		rec := post(NewHandler(duplicateStore{}), readEvent(t, "payment_succeeded.json"))
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Body too large", func(t *testing.T) {
		t.Parallel()

		h, _ := setup(t)
		rec := post(h, []byte(`{"id": "`+strings.Repeat("a", maxBodyBytes)+`"}`))
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "invalid JSON body", strings.TrimSpace(rec.Body.String()))
	})
}
