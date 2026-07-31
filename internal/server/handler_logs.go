package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/eraiza0816/switch-dashboard/internal/config"
)

func (s *Server) handleAPIGetLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.LogBuffer.Get())
}

func (s *Server) handleAPISetLogLevel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Level string `json:"level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var lvl slog.Level
	switch strings.ToUpper(req.Level) {
	case "DEBUG":
		lvl = slog.LevelDebug
	case "INFO":
		lvl = slog.LevelInfo
	case "WARN":
		lvl = slog.LevelWarn
	case "ERROR":
		lvl = slog.LevelError
	default:
		http.Error(w, "invalid level, use DEBUG/INFO/WARN/ERROR", http.StatusBadRequest)
		return
	}

	s.LogBuffer.SetLevel(lvl)
	slog.SetLogLoggerLevel(lvl)
	s.Logger.Info("log level changed", "level", req.Level)

	if s.ConfigPath != "" {
		cfg, err := config.Load(s.ConfigPath)
		if err == nil {
			cfg.Settings.LogLevel = req.Level
			cfg.Save()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "level": req.Level})
}

func (s *Server) handleAPIClearLogs(w http.ResponseWriter, r *http.Request) {
	s.LogBuffer.Clear()
	s.Logger.Info("logs cleared")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleAPIDownloadLogs(w http.ResponseWriter, r *http.Request) {
	entries := s.LogBuffer.Get()
	var sb strings.Builder
	for _, e := range entries {
		sb.WriteString(e.Time)
		sb.WriteString(" [")
		sb.WriteString(e.Level)
		sb.WriteString("] ")
		sb.WriteString(e.Message)
		if len(e.Attrs) > 0 {
			sb.WriteString(" ")
			enc := json.NewEncoder(&sb)
			enc.Encode(e.Attrs)
		}
		sb.WriteString("\n")
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=switch-dashboard.log")
	w.Write([]byte(sb.String()))
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:           s.Config.Title(),
		Version:         Version,
		CurrentLogLevel: strings.ToUpper(s.LogBuffer.Level().String()),
	}
	s.renderTemplate(w, "logs.html", data)
}


