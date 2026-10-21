//go:build fixed

package webhook

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_ServeHTTP_Fixed(t *testing.T) {
	t.Parallel()

	t.Run("Body too large", func(t *testing.T) {
		t.Parallel()

		h, _ := setup(t)
		rec := post(h, []byte(`{"id": "`+strings.Repeat("a", maxBodyBytes)+`"}`))
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "invalid JSON body", strings.TrimSpace(rec.Body.String()))
	})
}
