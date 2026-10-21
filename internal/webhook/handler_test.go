package webhook

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ainsleyclark/hookrelay/internal/store"
)

func setup(t *testing.T) (*Handler, *store.Memory) {
	t.Helper()

	st := store.NewMemory()
	return NewHandler(st), st
}

func readEvent(t *testing.T, name string) []byte {
	t.Helper()

	body, err := os.ReadFile(filepath.Join("testdata", "events", name))
	require.NoError(t, err)
	return body
}

// duplicateStore finds nothing on Get but reports every Save as a
// duplicate, as if another delivery of the same event won the race.
type duplicateStore struct{}

func (duplicateStore) Get(context.Context, string) (store.Event, error) {
	return store.Event{}, store.ErrNotFound
}

func (duplicateStore) Save(context.Context, store.Event) error {
	return store.ErrDuplicate
}

func post(h http.Handler, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", bytes.NewReader(body))
	req.SetPathValue("provider", "stripe")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	t.Run("New event saved", func(t *testing.T) {
		t.Parallel()

		h, st := setup(t)
		rec := post(h, readEvent(t, "payment_succeeded.json"))
		require.Equal(t, http.StatusAccepted, rec.Code)

		got, err := st.Get(t.Context(), "evt_1A2b3C4d5E6f7G8h9I0jK1lM")
		require.NoError(t, err)
		assert.Equal(t, "stripe", got.Provider)
	})

	t.Run("Duplicate event acknowledged", func(t *testing.T) {
		t.Parallel()

		h, _ := setup(t)
		body := readEvent(t, "payment_succeeded.json")
		require.Equal(t, http.StatusAccepted, post(h, body).Code)

		rec := post(h, body)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Racing duplicate acknowledged", func(t *testing.T) {
		t.Parallel()

		rec := post(NewHandler(duplicateStore{}), readEvent(t, "payment_succeeded.json"))
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Bad requests rejected", func(t *testing.T) {
		t.Parallel()

		type response struct {
			code int
			msg  string
		}

		tt := map[string]struct {
			input []byte
			want  response
		}{
			"Invalid event ID": {
				input: readEvent(t, "invalid_id.json"),
				want:  response{code: http.StatusBadRequest, msg: `invalid event id "evt_123"`},
			},
			"Malformed JSON": {
				input: []byte(`{"id":`),
				want:  response{code: http.StatusBadRequest, msg: "invalid JSON body"},
			},
			"Empty body": {
				input: []byte(""),
				want:  response{code: http.StatusBadRequest, msg: "invalid JSON body"},
			},
		}

		for name, test := range tt {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				h, _ := setup(t)
				rec := post(h, test.input)

				got := response{code: rec.Code, msg: strings.TrimSpace(rec.Body.String())}
				assert.Equal(t, test.want, got)
			})
		}
	})
}

func BenchmarkHandler(b *testing.B) {
	h := NewHandler(store.NewMemory())
	payloads := make([][]byte, b.N)
	for i := range payloads {
		payloads[i] = fmt.Appendf(nil,
			`{"id": "evt_%024d", "type": "payment.succeeded", "data": {"amount": 2500, "currency": "gbp"}}`, i)
	}

	b.ResetTimer()
	for i := range b.N {
		post(h, payloads[i])
	}
}
