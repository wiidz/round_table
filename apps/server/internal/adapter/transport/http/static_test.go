package httptransport

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegisterWebUI_spaFallback(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<html>app</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	assets := filepath.Join(root, "assets")
	if err := os.Mkdir(assets, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assets, "app.js"), []byte("console.log('ok')"), 0o644); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	if err := RegisterWebUI(mux, root); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/settings", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "app") {
		t.Fatalf("spa fallback status=%d body=%q", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "ok") {
		t.Fatalf("asset status=%d body=%q", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/meetings", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("api passthrough status=%d", rec.Code)
	}
}
