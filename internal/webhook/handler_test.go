package webhook

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ainsleyclark/hookrelay/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func post(h http.Handler, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", bytes.NewReader(body))
	req.SetPathValue("provider", "stripe")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHandler_ServeHTTP_Fixed(t *testing.T) {
	t.Parallel()

	t.Run("Body too large", func(t *testing.T) {
		t.Parallel()

		h, _ := setup(t)
		rec := post(h, []byte(`{"id": "`+strings.Repeat("a", maxBodyBytes)+`"}`))
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "invalid JSON body", strings.TrimSpace(rec.Body.String()))
	})
}
