package webhook

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ainsleyclark/hookrelay/internal/store"
)

func BenchmarkHandler(b *testing.B) {
	h := NewHandler(store.NewMemory())
	i := 0
	for b.Loop() {
		i++
		body := fmt.Sprintf(`{"id":"evt_%020d","type":"payment.succeeded"}`, i)
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		h.ServeHTTP(httptest.NewRecorder(), req)
	}
}
