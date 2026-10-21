package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"
)

const tolerance = 5 * time.Minute

var (
	errMalformedHeader = errors.New("malformed signature header")
	errExpired         = errors.New("signature expired")
	errMismatch        = errors.New("invalid signature")
)

// HMACVerifier accepts headers of the form "t=<unix>,v1=<signature>", where
// the signature is the hex HMAC-SHA256 of "<t>.<body>" keyed with Secret.
// Timestamps older than five minutes are rejected.
type HMACVerifier struct {
	Secret []byte
}

var _ Verifier = HMACVerifier{}

func (v HMACVerifier) Verify(header string, body []byte) error {
	ts, sig, err := parseHeader(header)
	if err != nil {
		return err
	}
	if time.Since(time.Unix(ts, 0)) > tolerance {
		return errExpired
	}
	if !hmac.Equal(sig, v.sign(ts, body)) {
		return errMismatch
	}
	return nil
}

func (v HMACVerifier) sign(ts int64, body []byte) []byte {
	mac := hmac.New(sha256.New, v.Secret)
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
