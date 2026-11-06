package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testSecret = []byte("test-secret")

func signHeader(secret []byte, ts int64, body string) string {
	t := strconv.FormatInt(ts, 10)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(t + "." + body))
	return "t=" + t + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

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

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "could not read request body", strings.TrimSpace(rec.Body.String()))
	})
}
