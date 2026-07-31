package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleAPIBackup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleAPIReboot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleAPIUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleAPIListBackups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func (s *Server) handleAPIDownloadBackup(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		http.Error(w, "missing filename", http.StatusBadRequest)
		return
	}
	backupDir := "./backup"
	filepath := filepath.Join(backupDir, filename)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	http.ServeFile(w, r, filepath)
}

func (s *Server) handleAPIDeleteBackup(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		http.Error(w, "missing filename", http.StatusBadRequest)
		return
	}
	backupDir := "./backup"
	filepath := filepath.Join(backupDir, filename)
	if err := os.Remove(filepath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
