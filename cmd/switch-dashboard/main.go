package main

import (
	"context"
	"flag"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/eraiza0816/switch-dashboard/internal/config"
	"github.com/eraiza0816/switch-dashboard/internal/history"
	"github.com/eraiza0816/switch-dashboard/internal/logbuf"
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

func main() {
	dataDir := flag.String("d", "", "data directory (default: ~/.local/share/switch-dashboard)")
	configPath := flag.String("c", "", "config file path (default: <data-dir>/config.json)")
	flag.Parse()

	dd := resolveDataDir(*dataDir)
	if err := os.MkdirAll(dd, 0755); err != nil {
		slog.Error("cannot create data directory", "dir", dd, "error", err)
		os.Exit(1)
	}

	cp := *configPath
	if cp == "" {
		cp = filepath.Join(dd, "config.json")
	}

	cfg, err := config.Load(cp)
	if err != nil {
		slog.Warn("cannot load config", "path", cp, "error", err)
		cfg = &config.Config{}
		cfg.SetDefaults()
	}

	logLevel := slog.LevelInfo
	if lvl := cfg.Settings.LogLevel; lvl != "" {
		switch strings.ToUpper(lvl) {
		case "DEBUG":
			logLevel = slog.LevelDebug
		case "INFO":
			logLevel = slog.LevelInfo
		case "WARN":
			logLevel = slog.LevelWarn
		case "ERROR":
			logLevel = slog.LevelError
		}
	}

	innerHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	logBuf := logbuf.New(innerHandler, 5000, logLevel)
	logger := slog.New(logBuf)
	slog.SetDefault(logger)

	logger.Info("starting switch-dashboard", "data_dir", dd, "config", cp, "log_level", logLevel.String())

	cache := server.NewCache()
	enabledCols := cfg.EnabledColumns
	if len(enabledCols) == 0 {
		enabledCols = []string{"port", "status", "speed", "packets", "bytes", "info", "notes"}
	}
	cfgProvider := &appConfig{
		title:          cfg.Title,
		refresh:        cfg.RefreshInterval,
		enabledColumns: enabledCols,
		version:        "0.1.0",
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

	reloadFn := func() error {
		cfg2, err := config.Load(cp)
		if err != nil {
			return err
		}
		logger.Info("config reloaded", "title", cfg2.Title, "switches", len(cfg2.ActiveSwitches()))
		server.SetActiveSwitchesGauge(len(cfg2.ActiveSwitches()))
		return nil
	}

	srv := server.NewServer(cache, cfgProvider, logger, logBuf, nil, mustStaticFS())
	srv.ConfigReload = reloadFn
	srv.HistoryStore = histStore
	if histStore != nil {
		srv.DuckDBReady = true
		server.SetDuckDBReadyGauge(true)
	} else {
		server.SetDuckDBReadyGauge(false)
	}
	server.SetActiveSwitchesGauge(len(cfg.ActiveSwitches()))
	srv.ClientHostsPath = filepath.Join(dd, "clients.json")
	srv.LayoutPositionsPath = filepath.Join(dd, "layout_positions.json")
	srv.OUI = oui.New()
	srv.DataDir = dd
	srv.ConfigPath = cp

	notes := cfg.Notes
	if notes == nil {
		notes = make(map[string]string)
	}

	var pollers []*poller.Poller
	for _, sw := range cfg.ActiveSwitches() {
		p := startPolling(cache, sw.IP, sw.Name, sw.Model, cfg.RefreshInterval, logger, notes, histStore)
		pollers = append(pollers, p)
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

			p := startPollingWithClient(cache, client, ip, sw.Name, sw.Model, cfg.RefreshInterval, logger, notes, histStore)
			pollers = append(pollers, p)
		}()
	}

	addr := os.Getenv("LISTEN")
	if addr == "" {
		addr = ":8081"
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	httpServer := &http.Server{Addr: addr, Handler: srv.Router}
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("starting server", "addr", addr, "data_dir", dd, "config", cp)
		if err := httpServer.ListenAndServe(); err != nil {
			serverErr <- err
		}
	}()

	// Signal/event loop
	for {
		select {
		case sig := <-sigCh:
			switch sig {
			case syscall.SIGHUP:
				logger.Info("reloading config")
				if err := reloadFn(); err != nil {
					logger.Warn("config reload failed", "error", err)
				}
			default:
				logger.Info("shutting down", "signal", sig)
				goto shutdown
			}
		case err := <-serverErr:
			logger.Error("server failed, exiting", "error", err)
			return
		}
	}

shutdown:

	for _, p := range pollers {
		p.Stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown error", "error", err)
	}

	if histStore != nil {
		if err := histStore.Close(); err != nil {
			logger.Error("history close error", "error", err)
		}
	}
	logger.Info("stopped")
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

func startPolling(cache *server.Cache, ip, name, model string, interval int, logger *slog.Logger, notes map[string]string, histStore *history.Store) *poller.Poller {
	p := poller.NewWithNotes(cache, ip, name, model, interval, notes, logger)
	if histStore != nil {
		p.SetHistoryStore(histStore)
	}
	p.Start()
	return p
}

func startPollingWithClient(cache *server.Cache, client *rtlplayground.Client, ip, name, model string, interval int, logger *slog.Logger, notes map[string]string, histStore *history.Store) *poller.Poller {
	p := poller.NewWithClientAndNotes(cache, client, ip, name, model, interval, notes, logger)
	if histStore != nil {
		p.SetHistoryStore(histStore)
	}
	p.Start()
	return p
}

func mustStaticFS() fs.FS {
	dir := "./static"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return os.DirFS(".")
}
