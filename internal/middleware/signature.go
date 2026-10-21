// Package middleware contains HTTP middleware for webhook routes.
package middleware

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
)

// SignatureHeader is the request header that carries the webhook signature.
const SignatureHeader = "Webhook-Signature"

const maxBodyBytes = 1 << 20

// Verifier checks a signature header against the raw request body.
type Verifier interface {
	Verify(header string, body []byte) error
}

// VerifySignature rejects requests whose SignatureHeader is not accepted
// by v. Accepted requests reach next with their body intact.
func VerifySignature(v Verifier, next http.Handler) http.Handler {
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
		if err := v.Verify(r.Header.Get(SignatureHeader), body); err != nil {
			slog.WarnContext(r.Context(), "rejected webhook signature", "error", err)
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(w, r)
	})
}
