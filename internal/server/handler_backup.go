package server

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleAPIBackup(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAPIReboot(w http.ResponseWriter, r *http.Request) {
	ip := chi.URLParam(r, "ip")
	client, err := s.clientFor(ip)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err := client.Reboot(); err != nil {
		s.writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAPIUpload(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAPIListBackups(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, []string{})
}

func (s *Server) handleAPIDownloadBackup(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		s.writeError(w, http.StatusBadRequest, "missing filename")
		return
	}
	backupDir := "./backup"
	filepath := filepath.Join(backupDir, filename)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		s.writeError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	http.ServeFile(w, r, filepath)
}

func (s *Server) handleAPIDeleteBackup(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		s.writeError(w, http.StatusBadRequest, "missing filename")
		return
	}
	backupDir := "./backup"
	filepath := filepath.Join(backupDir, filename)
	if err := os.Remove(filepath); err != nil {
		if os.IsNotExist(err) {
			s.writeError(w, http.StatusNotFound, "not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
