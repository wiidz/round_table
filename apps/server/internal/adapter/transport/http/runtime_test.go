package httptransport

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"round_table/apps/server/internal/platform/config"
	"round_table/apps/server/internal/platform/procstats"
)

func TestHandleGetRuntime(t *testing.T) {
	procstats.MarkServerStarted()
	cfg := config.Load()
	configSvc, err := config.NewService(nil)
	if err != nil {
		t.Fatal(err)
	}
	h, err := NewHandler(cfg, nil, nil, configSvc, nil)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/system/runtime", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"server"`) {
		t.Fatalf("body=%s", body)
	}
	if !strings.Contains(body, `"memory_bytes"`) {
		t.Fatalf("body=%s", body)
	}
	if !strings.Contains(body, `"listen_addr"`) {
		t.Fatalf("body=%s", body)
	}
}
