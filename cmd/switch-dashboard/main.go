package main

import (
	"flag"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/eraiza0816/switch-dashboard/internal/config"
	"github.com/eraiza0816/switch-dashboard/internal/history"
	"github.com/eraiza0816/switch-dashboard/internal/oui"
	"github.com/eraiza0816/switch-dashboard/internal/poller"
	"github.com/eraiza0816/switch-dashboard/internal/rtlplayground"
	"github.com/eraiza0816/switch-dashboard/internal/server"
)

type appConfig struct {
	title              string
	refresh            int
	enabledColumns     []string
	gridColumns        string
	portsWrapThreshold int
	columnWidths       map[string]int
	columnOrder        []string
	version            string
}

func (a *appConfig) Title() string                    { return a.title }
func (a *appConfig) RefreshInterval() int             { return a.refresh }
func (a *appConfig) EnabledColumns() []string         { return a.enabledColumns }
func (a *appConfig) GridColumns() string              { return a.gridColumns }
func (a *appConfig) PortsWrapThreshold() int          { return a.portsWrapThreshold }
func (a *appConfig) ColumnWidths() map[string]int     { return a.columnWidths }
func (a *appConfig) ColumnOrder() []string            { return a.columnOrder }
func (a *appConfig) Version() string                  { return a.version }

type appLogger struct {
	logger *slog.Logger
}

func (a *appLogger) Info(msg string, args ...any)  { a.logger.Info(msg, args...) }
func (a *appLogger) Warn(msg string, args ...any)  { a.logger.Warn(msg, args...) }
func (a *appLogger) Error(msg string, args ...any) { a.logger.Error(msg, args...) }
func (a *appLogger) Debug(msg string, args ...any) { a.logger.Debug(msg, args...) }

func main() {
	dataDir := flag.String("d", "", "data directory (default: ~/.local/share/switch-dashboard)")
	configPath := flag.String("c", "", "config file path (default: <data-dir>/config.json)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	al := &appLogger{logger: logger}

	dd := resolveDataDir(*dataDir)
	if err := os.MkdirAll(dd, 0755); err != nil {
		logger.Error("cannot create data directory", "dir", dd, "error", err)
		os.Exit(1)
	}

	cp := *configPath
	if cp == "" {
		cp = filepath.Join(dd, "config.json")
	}

	cfg, err := config.Load(cp)
	if err != nil {
		logger.Warn("cannot load config", "path", cp, "error", err)
		cfg = &config.Config{}
		cfg.SetDefaults()
	}

	cache := server.NewCache()
	enabledCols := cfg.EnabledColumns
	if len(enabledCols) == 0 {
		enabledCols = []string{"port", "status", "speed", "packets", "bytes", "info", "notes"}
	}
	cfgProvider := &appConfig{
		title:              cfg.Title,
		refresh:            cfg.RefreshInterval,
		enabledColumns:     enabledCols,
		version:            "0.1.0",
	}

	histPath := filepath.Join(dd, "history.duckdb")
	histStore, err := history.NewStore(histPath)
	if err != nil {
		logger.Warn("cannot open history store, history will be disabled", "error", err)
	}
	if histStore != nil {
		if err := histStore.Retain(365 * 24 * time.Hour); err != nil {
			logger.Warn("history retention cleanup failed", "error", err)
		}
		go func() {
			for {
				time.Sleep(24 * time.Hour)
				if err := histStore.Retain(365 * 24 * time.Hour); err != nil {
					logger.Warn("history retention cleanup failed", "error", err)
				}
			}
		}()
	}

	srv := server.NewServer(cache, cfgProvider, al, nil, mustStaticFS())
	srv.HistoryStore = histStore
	srv.ClientHostsPath = filepath.Join(dd, "clients.json")
	srv.LayoutPositionsPath = filepath.Join(dd, "layout_positions.json")
	srv.OUI = oui.New()
	srv.DataDir = dd
	srv.ConfigPath = cp

	notes := cfg.Notes
	if notes == nil {
		notes = make(map[string]string)
	}

	for _, sw := range cfg.ActiveSwitches() {
		startPolling(cache, sw.IP, sw.Name, sw.Model, cfg.RefreshInterval, logger, notes, histStore)
	}

	for _, sw := range cfg.ActiveSwitches() {
		sw := sw
		go func() {
			ip := sw.IP
			password := sw.Password
			logger.Info("connecting to switch", "ip", ip)

			client, err := rtlplayground.New(ip, password)
			if err != nil {
				logger.Warn("falling back to mock data", "ip", ip, "error", err)
				return
			}

			info, err := client.ScrapeInformation()
			if err != nil {
				logger.Warn("initial scrape failed, keeping mock", "ip", ip, "error", err)
				return
			}
			logger.Info("connected, switching to live data", "ip", ip, "model", info.HWVer)

			startPollingWithClient(cache, client, ip, sw.Name, sw.Model, cfg.RefreshInterval, logger, notes, histStore)
		}()
	}

	addr := os.Getenv("LISTEN")
	if addr == "" {
		addr = ":8081"
	}
	logger.Info("starting server", "addr", addr, "data_dir", dd, "config", cp)
	if err := http.ListenAndServe(addr, srv.Router); err != nil {
		logger.Error("server failed", "error", err)
	}
	if histStore != nil {
		histStore.Close()
	}
}

func resolveDataDir(flagVal string) string {
	if flagVal != "" {
		return flagVal
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".local", "share", "switch-dashboard")
	}
	return "./data"
}

func startPolling(cache *server.Cache, ip, name, model string, interval int, logger *slog.Logger, notes map[string]string, histStore *history.Store) {
	p := poller.NewWithNotes(cache, ip, name, model, interval, notes)
	if histStore != nil {
		p.SetHistoryStore(histStore)
	}
	p.Start()
}

func startPollingWithClient(cache *server.Cache, client *rtlplayground.Client, ip, name, model string, interval int, logger *slog.Logger, notes map[string]string, histStore *history.Store) {
	p := poller.NewWithClientAndNotes(cache, client, ip, name, model, interval, notes)
	if histStore != nil {
		p.SetHistoryStore(histStore)
	}
	p.Start()
}

func mustStaticFS() fs.FS {
	dir := "./static"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return os.DirFS(".")
}
