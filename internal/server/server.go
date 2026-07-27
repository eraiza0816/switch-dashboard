package server

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/eraiza0816/switch-dashboard/internal/history"
	"github.com/eraiza0816/switch-dashboard/internal/oui"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var Version = "0.1.0"

type Server struct {
	Router          *chi.Mux
	Cache           *Cache
	Config          ConfigProvider
	Logger          Logger
	tmpl            *template.Template
	staticFS        fs.FS
	ClientHosts        map[string]string
	ClientHostsPath    string
	clientHostsMu      sync.RWMutex
	LayoutPositions    map[string]map[string]float64
	LayoutPositionsPath string
	layoutPositionsMu  sync.RWMutex
	HistoryStore       *history.Store
	OUI               *oui.DB
	DataDir           string
	ConfigPath        string
}

func (s *Server) ClientHost(mac string) string {
	s.clientHostsMu.RLock()
	defer s.clientHostsMu.RUnlock()
	if s.ClientHosts == nil {
		return ""
	}
	return s.ClientHosts[mac]
}

func (s *Server) loadClientHosts() {
	if s.ClientHostsPath == "" {
		return
	}
	data, err := os.ReadFile(s.ClientHostsPath)
	if err != nil {
		return
	}
	var hosts map[string]string
	if err := json.Unmarshal(data, &hosts); err != nil {
		return
	}
	s.clientHostsMu.Lock()
	s.ClientHosts = hosts
	s.clientHostsMu.Unlock()
}

func (s *Server) saveClientHosts() {
	if s.ClientHostsPath == "" {
		return
	}
	s.clientHostsMu.RLock()
	data, err := json.MarshalIndent(s.ClientHosts, "", "  ")
	s.clientHostsMu.RUnlock()
	if err != nil {
		return
	}
	os.WriteFile(s.ClientHostsPath, data, 0644)
}

func (s *Server) loadLayoutPositions() {
	if s.LayoutPositionsPath == "" {
		return
	}
	data, err := os.ReadFile(s.LayoutPositionsPath)
	if err != nil {
		return
	}
	var positions map[string]map[string]float64
	if err := json.Unmarshal(data, &positions); err != nil {
		return
	}
	s.layoutPositionsMu.Lock()
	s.LayoutPositions = positions
	s.layoutPositionsMu.Unlock()
}

func (s *Server) saveLayoutPositions() {
	if s.LayoutPositionsPath == "" {
		return
	}
	s.layoutPositionsMu.RLock()
	data, err := json.MarshalIndent(s.LayoutPositions, "", "  ")
	s.layoutPositionsMu.RUnlock()
	if err != nil {
		return
	}
	os.WriteFile(s.LayoutPositionsPath, data, 0644)
}

type ConfigProvider interface {
	Title() string
	RefreshInterval() int
	EnabledColumns() []string
	GridColumns() string
	PortsWrapThreshold() int
	ColumnWidths() map[string]int
	ColumnOrder() []string
	Version() string
}

type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

type simpleConfig struct {
	title              string
	refresh            int
	enabledColumns     []string
	gridColumns        string
	portsWrapThreshold int
	columnWidths       map[string]int
	columnOrder        []string
	version            string
}

func (s *simpleConfig) Title() string                    { return s.title }
func (s *simpleConfig) RefreshInterval() int             { return s.refresh }
func (s *simpleConfig) EnabledColumns() []string         { return s.enabledColumns }
func (s *simpleConfig) GridColumns() string              { return s.gridColumns }
func (s *simpleConfig) PortsWrapThreshold() int          { return s.portsWrapThreshold }
func (s *simpleConfig) ColumnWidths() map[string]int     { return s.columnWidths }
func (s *simpleConfig) ColumnOrder() []string            { return s.columnOrder }
func (s *simpleConfig) Version() string                  { return s.version }

func NewServer(cache *Cache, cfg ConfigProvider, logger Logger, tmplFS fs.FS, staticFS fs.FS) *Server {
	s := &Server{
		Router:      chi.NewRouter(),
		Cache:       cache,
		Config:      cfg,
		Logger:      logger,
		staticFS:    staticFS,
			ClientHosts:     make(map[string]string),
		LayoutPositions: make(map[string]map[string]float64),
	}

	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)
	s.Router.Use(middleware.RealIP)

	s.tmpl = loadTemplatesWithFS(tmplFS)
	s.registerRoutes()
	s.loadClientHosts()
	s.loadLayoutPositions()
	return s
}

func loadTemplatesWithFS(tmplFS fs.FS) *template.Template {
	funcs := templateFuncs()
	if tmplFS != nil {
		return template.Must(template.New("").Funcs(funcs).ParseFS(tmplFS, "*.html"))
	}
	// Fall back to disk
	dir := "templates"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		placeholder := `<!DOCTYPE html><html><body>{{ .Title }}</body></html>`
		t := template.New("index.html").Funcs(funcs)
		template.Must(t.Parse(placeholder))
		template.Must(t.New("config.html").Parse(placeholder))
		template.Must(t.New("backups.html").Parse(placeholder))
		template.Must(t.New("logs.html").Parse(placeholder))
		template.Must(t.New("api_docs.html").Parse(placeholder))
		template.Must(t.New("map.html").Parse(placeholder))
		return t
	}
	glob := filepath.Join(dir, "*.html")
	return template.Must(template.New("").Funcs(funcs).ParseGlob(glob))
}

func (s *Server) registerRoutes() {
	s.Router.Get("/", s.handleDashboard)
	s.Router.Get("/config", s.handleConfig)
	s.Router.Post("/config", s.handleConfigSave)
	s.Router.Get("/map", s.handleMap)
	s.Router.Get("/backups", s.handleBackups)
	s.Router.Get("/logs", s.handleLogs)
	s.Router.Get("/api-docs", s.handleAPIDocs)

	s.Router.Route("/api", func(r chi.Router) {
		r.Get("/switches", s.handleAPISwitches)
		r.Get("/switches/{ip}/sfp", s.handleAPISwitchSFP)
		r.Get("/switches/{ip}/transceiver", s.handleAPISwitchSFP)
		r.Post("/switches/{ip}/refresh_mac", s.handleAPIRefreshMAC)
		r.Post("/switches/{ip}/backup", s.handleAPIBackup)
		r.Post("/switches/{ip}/reboot", s.handleAPIReboot)
		r.Post("/switches/{ip}/upload", s.handleAPIUpload)
		r.Post("/switches/{ip}/cmd", s.handleAPISwitchCmd)

		r.Get("/speeds", s.handleAPISpeeds)
		r.Get("/history", s.handleAPIHistory)
		r.Post("/notes", s.handleAPINotes)
		r.Post("/reset", s.handleAPIReset)
		r.Get("/topology", s.handleAPITopology)

		r.Get("/settings", s.handleAPIGetSettings)
		r.Post("/settings", s.handleAPISaveSettings)
		r.Get("/config/settings", s.handleAPIGetConfigSettings)
		r.Post("/config/settings", s.handleAPISaveConfigSettings)

		r.Get("/logs", s.handleAPIGetLogs)
		r.Post("/logs/level", s.handleAPISetLogLevel)
		r.Post("/logs/clear", s.handleAPIClearLogs)
		r.Get("/logs/download", s.handleAPIDownloadLogs)

		r.Get("/vendors", s.handleAPIGetVendors)
		r.Post("/vendors", s.handleAPISaveVendors)
		r.Post("/vendors/update_oui", s.handleAPIUpdateOUI)
		r.Post("/clients/update_host", s.handleAPIUpdateHost)
		r.Get("/layout_positions", s.handleAPIGetLayoutPositions)
		r.Post("/layout_positions", s.handleAPISaveLayoutPositions)

		r.Get("/openapi.json", s.handleOpenAPI)
	})

	s.Router.Route("/api/backups", func(r chi.Router) {
		r.Get("/", s.handleAPIListBackups)
		r.Get("/{filename}/download", s.handleAPIDownloadBackup)
		r.Delete("/{filename}", s.handleAPIDeleteBackup)
	})

	s.Router.Get("/backups", s.handleBackups)

	if s.staticFS != nil {
		fileServer := http.FileServer(http.FS(s.staticFS))
		s.Router.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = path.Join("/static", chi.URLParam(r, "*"))
			fileServer.ServeHTTP(w, r)
		})
	}
}

func (s *Server) renderTemplate(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if s.tmpl == nil {
		w.Write([]byte("<html><body>template not loaded</body></html>"))
		return
	}
	if err := s.tmpl.ExecuteTemplate(w, name, data); err != nil {
		s.Logger.Error("template error", "name", name, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
