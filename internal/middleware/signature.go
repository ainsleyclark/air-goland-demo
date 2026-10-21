package middleware

import (
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

const tolerance = 5 * time.Minute

// VerifySignature rejects requests that aren't signed with secret or are
// more than five minutes old.
func VerifySignature(secret []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// ISSUE: 	The body is read with no size limit, and the read error is discarded.
		// FIX: 	Code Review Skill should flag the missing http.MaxBytesReader and the unchecked error.
		body, _ := io.ReadAll(r.Body)

		// ISSUE: 	Verification is hard-wired to one HMAC scheme and one secret, so a second provider means rewriting this middleware.
		// FIX: 	Refactoring Skill should pull verification behind a Verifier interface with one HMAC implementation.
		if err := verify(secret, r.Header.Get(SignatureHeader), body); err != nil {
			slog.WarnContext(ctx, "rejected webhook signature", "error", err)
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		// ISSUE: 	The body was read above and never put back, so the handler decodes an empty body and every real request fails with 400.
		// FIX: 	Debugger Skill should trace the empty body from the handler back here and restore r.Body with io.NopCloser(bytes.NewReader(body)).
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

	// ISSUE: 	== isn't constant time, so response timing can leak how much of the signature matched.
	// FIX: 	Diff feedback on Agent 2: "use hmac.Equal, not ==." Code Review Skill should flag it too.
	if sig != hex.EncodeToString(mac.Sum(nil)) {
		return errors.New("signature mismatch")
	}

	return nil
}
