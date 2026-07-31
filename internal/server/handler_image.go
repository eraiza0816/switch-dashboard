package server

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

// handleAPISwitchImage serves a model-specific device image for a switch card.
// Images are looked up in <data-dir>/device-templates/<model>.<ext>, falling
// back to the bundled logo.
func (s *Server) handleAPISwitchImage(w http.ResponseWriter, r *http.Request) {
	ip := chi.URLParam(r, "ip")
	sw := s.Cache.GetSwitch(ip)
	model := ""
	if sw != nil {
		model = sw.Model
	}

	if model != "" && s.DeviceTemplatesDir != "" {
		for _, ext := range []string{"png", "jpg", "jpeg"} {
			candidate := filepath.Join(s.DeviceTemplatesDir, model+"."+ext)
			if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
				http.ServeFile(w, r, candidate)
				return
			}
		}
	}

	// Fallback to bundled logo.
	if s.staticFS != nil {
		f, err := s.staticFS.Open("static/logo.png")
		if err == nil {
			defer f.Close()
			content, err := io.ReadAll(f)
			if err == nil {
				w.Header().Set("Content-Type", "image/png")
				w.Header().Set("Cache-Control", "public, max-age=3600")
				w.Write(content)
				return
			}
		}
	}

	s.writeError(w, http.StatusNotFound, "image not found")
}
