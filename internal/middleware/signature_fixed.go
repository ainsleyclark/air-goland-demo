//go:build fixed

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

// Verifier checks a signature header against the raw request body.
// Each provider's signing scheme is one implementation.
type Verifier interface {
	Verify(header string, body []byte) error
}

// VerifySignature rejects requests that aren't signed with secret or are
// more than five minutes old.
func VerifySignature(secret []byte, next http.Handler) http.Handler {
	return VerifyWith(HMACVerifier{Secret: secret}, next)
}

// VerifyWith rejects requests whose signature v doesn't accept. Accepted
// requests reach next with their body intact.
func VerifyWith(v Verifier, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
		if err != nil {
			slog.WarnContext(ctx, "reading webhook body", "error", err)
			if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
				http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			} else {
				http.Error(w, "could not read request body", http.StatusBadRequest)
			}
			return
		}

		if err = v.Verify(r.Header.Get(SignatureHeader), body); err != nil {
			slog.WarnContext(ctx, "rejected webhook signature", "error", err)
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(w, r)
	})
}

// HMACVerifier accepts "t=<unix>,v1=<signature>" headers, where the
// signature is the hex HMAC-SHA256 of "<t>.<body>" keyed with Secret.
type HMACVerifier struct {
	Secret []byte
}

// Verify implements Verifier.
func (v HMACVerifier) Verify(header string, body []byte) error {
	var ts int64
	var sig string
	if _, err := fmt.Sscanf(header, "t=%d,v1=%s", &ts, &sig); err != nil {
		return fmt.Errorf("parsing header: %w", err)
	}
	if time.Since(time.Unix(ts, 0)) > tolerance {
		return errors.New("signature expired")
	}

	mac := hmac.New(sha256.New, v.Secret)
	mac.Write([]byte(strconv.FormatInt(ts, 10) + "."))
	mac.Write(body)
	if !hmac.Equal([]byte(sig), []byte(hex.EncodeToString(mac.Sum(nil)))) {
		return errors.New("signature mismatch")
	}

	return nil
}
