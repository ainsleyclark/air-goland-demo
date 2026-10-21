// Package middleware contains HTTP middleware for webhook routes.
package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SignatureHeader holds "t=<unix>,v1=<signature>", where the signature is
// the hex HMAC-SHA256 of "<t>.<body>".
const SignatureHeader = "Webhook-Signature"

const (
	tolerance    = 5 * time.Minute
	maxBodyBytes = 1 << 20
)

var errMalformedHeader = errors.New("malformed signature header")

// VerifySignature rejects requests whose SignatureHeader was not produced
// with secret, or whose timestamp is older than five minutes.
func VerifySignature(secret []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ts, sig, err := parseHeader(r.Header.Get(SignatureHeader))
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if time.Since(time.Unix(ts, 0)) > tolerance {
			http.Error(w, "signature expired", http.StatusUnauthorized)
			return
		}
		if !hmac.Equal(sig, sign(secret, ts, body)) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(w, r)
	})
}

func sign(secret []byte, ts int64, body []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(strconv.FormatInt(ts, 10) + "."))
	mac.Write(body)
	return mac.Sum(nil)
}

func parseHeader(header string) (ts int64, sig []byte, err error) {
	for part := range strings.SplitSeq(header, ",") {
		key, value, _ := strings.Cut(strings.TrimSpace(part), "=")
		switch key {
		case "t":
			if ts, err = strconv.ParseInt(value, 10, 64); err != nil {
				return 0, nil, errMalformedHeader
			}
		case "v1":
			if sig, err = hex.DecodeString(value); err != nil {
				return 0, nil, errMalformedHeader
			}
		}
	}
	if ts == 0 || len(sig) == 0 {
		return 0, nil, errMalformedHeader
	}
	return ts, sig, nil
}
