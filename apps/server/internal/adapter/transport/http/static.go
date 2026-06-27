package httptransport

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// RegisterWebUI serves a Vite/React production build (SPA fallback to index.html).
func RegisterWebUI(mux *http.ServeMux, root string) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return fmt.Errorf("web root required")
	}
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("web root: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("web root is not a directory: %s", root)
	}

	handler := spaHandler{root: root}
	mux.Handle("GET /{$}", handler)
	mux.Handle("GET /{path...}", handler)
	return nil
}

type spaHandler struct {
	root string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api" {
		http.NotFound(w, r)
		return
	}

	rel := strings.TrimPrefix(r.URL.Path, "/")
	if rel == "" {
		h.serveFile(w, r, "index.html")
		return
	}

	clean := filepath.Clean(rel)
	if clean == "." || strings.HasPrefix(clean, "..") {
		http.NotFound(w, r)
		return
	}

	target := filepath.Join(h.root, clean)
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		h.serveFile(w, r, clean)
		return
	}

	h.serveFile(w, r, "index.html")
}

func (h spaHandler) serveFile(w http.ResponseWriter, r *http.Request, name string) {
	path := filepath.Join(h.root, filepath.Clean(name))
	if !strings.HasPrefix(path, h.root) {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, path)
}
