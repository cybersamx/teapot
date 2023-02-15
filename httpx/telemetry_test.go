package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleHealthCheck(t *testing.T) {
	ctx, cancel := newTestContext(t)
	defer cancel()

	server := newTestServer(t, ctx)

	record := httptest.NewRecorder()
	server.router.ServeHTTP(record, httptest.NewRequest(http.MethodGet, "/health", nil))
	assert.Equal(t, http.StatusOK, record.Code)
	assert.JSONEq(t, `{"status": "OK"}`, record.Body.String())
}
