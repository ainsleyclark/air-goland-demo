package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ainsleyclark/hookrelay/internal/middleware"
	"github.com/ainsleyclark/hookrelay/internal/store"
)

func signHeader(secret []byte, ts int64, body []byte) string {
	t := strconv.FormatInt(ts, 10)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(t + "." + string(body)))
	return "t=" + t + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

func TestWebhookEndToEnd(t *testing.T) {
	secret := []byte("test-secret")
	st := store.NewMemory()
	srv := httptest.NewServer(routes(st, middleware.HMACVerifier{Secret: secret}))
	defer srv.Close()

	body, err := os.ReadFile("../../testdata/events/payment_succeeded.json")
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []int{http.StatusAccepted, http.StatusOK} {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/webhooks/stripe", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set(middleware.SignatureHeader, signHeader(secret, time.Now().Unix(), body))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != want {
			t.Errorf("status = %d, want %d", resp.StatusCode, want)
		}
	}

	ev, err := st.Get(t.Context(), "evt_1A2b3C4d5E6f7G8h9I0jK1lM")
	if err != nil {
		t.Fatal(err)
	}
	if ev.Provider != "stripe" {
		t.Errorf("provider = %q, want %q", ev.Provider, "stripe")
	}
}
