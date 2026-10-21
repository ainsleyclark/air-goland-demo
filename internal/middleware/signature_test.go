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
)

var testSecret = []byte("test-secret")

func signHeader(secret []byte, ts int64, body string) string {
	t := strconv.FormatInt(ts, 10)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(t + "." + body))
	return "t=" + t + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifySignature(t *testing.T) {
	const body = `{"id": "evt_1A2b3C4d5E6f7G8h9I0jK1lM", "type": "payment.succeeded"}`
	now := time.Now().Unix()

	tests := []struct {
		name   string
		header string
		body   string
		want   int
	}{
		{"valid", signHeader(testSecret, now, body), body, http.StatusOK},
		{"tampered body", signHeader(testSecret, now, body), body + " ", http.StatusUnauthorized},
		{"wrong secret", signHeader([]byte("other"), now, body), body, http.StatusUnauthorized},
		{"expired", signHeader(testSecret, now-600, body), body, http.StatusUnauthorized},
		{"malformed header", "v1=abc", body, http.StatusUnauthorized},
		{"missing header", "", body, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []byte
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got, _ = io.ReadAll(r.Body)
			})
			req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", strings.NewReader(tt.body))
			req.Header.Set(SignatureHeader, tt.header)
			rec := httptest.NewRecorder()

			VerifySignature(HMACVerifier{Secret: testSecret}, next).ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
			if tt.want == http.StatusOK && string(got) != tt.body {
				t.Errorf("next read body %q, want %q", got, tt.body)
			}
		})
	}
}
