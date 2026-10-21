package webhook

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ainsleyclark/hookrelay/internal/store"
)

func readEvent(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "events", name))
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func post(h http.Handler, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", bytes.NewReader(body))
	req.SetPathValue("provider", "stripe")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHandler_ValidEvent(t *testing.T) {
	st := store.NewMemory()

	rec := post(NewHandler(st), readEvent(t, "payment_succeeded.json"))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}

	ev, err := st.Get(t.Context(), "evt_1A2b3C4d5E6f7G8h9I0jK1lM")
	if err != nil {
		t.Fatal(err)
	}
	if ev.Provider != "stripe" {
		t.Errorf("provider = %q, want %q", ev.Provider, "stripe")
	}
}

func TestHandler_Duplicate(t *testing.T) {
	h := NewHandler(store.NewMemory())
	body := readEvent(t, "payment_succeeded.json")

	post(h, body)
	if rec := post(h, body); rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHandler_InvalidID(t *testing.T) {
	rec := post(NewHandler(store.NewMemory()), readEvent(t, "invalid_id.json"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandler_BadBody(t *testing.T) {
	huge := `{"id": "` + strings.Repeat("a", maxBodyBytes) + `"}`

	tests := []struct {
		name string
		body string
		code int
		msg  string
	}{
		{"malformed", `{"id":`, http.StatusBadRequest, "invalid JSON body"},
		{"empty", "", http.StatusBadRequest, "invalid JSON body"},
		{"too large", huge, http.StatusRequestEntityTooLarge, "request body too large"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := post(NewHandler(store.NewMemory()), []byte(tt.body))

			if rec.Code != tt.code {
				t.Errorf("status = %d, want %d", rec.Code, tt.code)
			}
			if got := strings.TrimSpace(rec.Body.String()); got != tt.msg {
				t.Errorf("body = %q, want %q", got, tt.msg)
			}
		})
	}
}

func BenchmarkHandler(b *testing.B) {
	h := NewHandler(store.NewMemory())
	payloads := make([][]byte, b.N)
	for i := range payloads {
		payloads[i] = fmt.Appendf(nil,
			`{"id": "evt_%024d", "type": "payment.succeeded", "data": {"amount": 2500, "currency": "gbp"}}`, i)
	}

	b.ResetTimer()
	for i := range b.N {
		post(h, payloads[i])
	}
}
