package httpserver

import (
 "log"
 "net/http"
 "net/http/httptest"
 "testing"
)

func TestHealth(t *testing.T) {
 s := New(":0", "0.1.0", log.Default())
 req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
 rec := httptest.NewRecorder()
 s.health(rec, req)
 if rec.Code != http.StatusOK { t.Fatalf("expected 200, got %d", rec.Code) }
}
