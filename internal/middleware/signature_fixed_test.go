//go:build fixed

package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifySignature_Fixed(t *testing.T) {
	t.Parallel()

	const body = `{"id": "evt_1A2b3C4d5E6f7G8h9I0jK1lM", "type": "payment.succeeded"}`

	t.Run("Body reaches next intact", func(t *testing.T) {
		t.Parallel()

		var got string
		next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			read, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			got = string(read)
		})
		req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", strings.NewReader(body))
		req.Header.Set(SignatureHeader, signHeader(testSecret, time.Now().Unix(), body))
		rec := httptest.NewRecorder()

		VerifySignature(testSecret, next).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, body, got)
	})

	t.Run("Body too large", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", strings.NewReader(strings.Repeat("a", maxBodyBytes+1)))
		rec := httptest.NewRecorder()

		VerifySignature(testSecret, http.NotFoundHandler()).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	})
}
