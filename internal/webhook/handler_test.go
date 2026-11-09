package webhook

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ainsleyclark/hookrelay/internal/store"
)

func BenchmarkHandler(b *testing.B) {
	h := NewHandler(store.NewMemory())
	b.ReportAllocs()
	i := 0

	for b.Loop() {
		i++
		body := fmt.Sprintf(`{"id":"evt_%020d","type":"payment.succeeded"}`, i)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusAccepted {
			b.Fatalf("got status %d, want %d", rec.Code, http.StatusAccepted)
		}
	}
}
