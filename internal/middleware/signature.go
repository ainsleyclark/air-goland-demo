package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// SignatureHeader is the request header that carries the webhook signature.
const SignatureHeader = "Webhook-Signature"

const (
	maxBodyBytes = 1 << 20
	tolerance    = 5 * time.Minute
)

// VerifySignature rejects requests that aren't signed with secret or are
// more than five minutes old.
func VerifySignature(secret []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
		if err != nil {
			slog.WarnContext(r.Context(), "reading webhook body", "error", err)
			if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
				http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			} else {
				http.Error(w, "could not read request body", http.StatusBadRequest)
			}
			return
		}
		if err := verify(secret, r.Header.Get(SignatureHeader), body); err != nil {
			slog.WarnContext(r.Context(), "rejected webhook signature", "error", err)
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(w, r)
	})
}

func verify(secret []byte, header string, body []byte) error {
	var ts int64
	var sig string
	if _, err := fmt.Sscanf(header, "t=%d,v1=%s", &ts, &sig); err != nil {
		return fmt.Errorf("parsing header: %w", err)
	}
	if time.Since(time.Unix(ts, 0)) > tolerance {
		return errors.New("signature expired")
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(strconv.FormatInt(ts, 10) + "."))
	mac.Write(body)
	if !hmac.Equal([]byte(sig), []byte(hex.EncodeToString(mac.Sum(nil)))) {
		return errors.New("signature mismatch")
	}

	return nil
}
