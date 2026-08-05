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
	"sync"
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
	mu      sync.RWMutex
	cfg     *config.Config
	version string
}

func (a *appConfig) applyConfig(cfg *config.Config) {
	a.mu.Lock()
	a.cfg = cfg
	a.mu.Unlock()
}

func (a *appConfig) Title() string { a.mu.RLock(); defer a.mu.RUnlock(); return a.cfg.Title }
func (a *appConfig) RefreshInterval() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.RefreshInterval
}
func (a *appConfig) EnabledColumns() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.EnabledColumns
}
func (a *appConfig) GridColumns() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.GridColumns
}
func (a *appConfig) PortsWrapThreshold() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.PortsWrapThreshold
}
func (a *appConfig) ColumnWidths() map[string]int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.ColumnWidths
}
func (a *appConfig) ColumnOrder() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.ColumnOrder
}
func (a *appConfig) Version() string { return a.version }
func (a *appConfig) Switches() []config.SwitchConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.Switches
}
func (a *appConfig) InfrastructureDevices() []config.InfraDevice {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.InfrastructureDevices
}
func (a *appConfig) UnmanagedSwitches() []config.UnmanagedSwitch {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.UnmanagedSwitches
}
func (a *appConfig) IgnoredMACs() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.Settings.IgnoredMACs
}

func main() {
	dataDir := flag.String("d", "", "data directory (default: ~/.local/share/switch-dashboard)")
	configPath := flag.String("c", "", "config file path (default: <data-dir>/config.json)")
	demo := flag.Bool("demo", false, "demo mode: seed mock data, no real switch required")
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
	cfg.DemoMode = *demo

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
	cfgProvider := &appConfig{version: "0.1.0"}
	cfgProvider.applyConfig(cfg)

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
		cfgProvider.applyConfig(cfg2)
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
	srv.ClientsPath = filepath.Join(dd, "clients.json")
	srv.LayoutPositionsPath = filepath.Join(dd, "layout_positions.json")
	srv.OUI = oui.New()
	srv.DataDir = dd
	srv.ConfigPath = cp
	srv.DeviceTypesPath = filepath.Join(dd, "device_types.yaml")
	srv.DeviceTemplatesDir = filepath.Join(dd, "device-templates")
	if err := os.MkdirAll(srv.DeviceTemplatesDir, 0755); err != nil {
		logger.Warn("cannot create device-templates dir", "dir", srv.DeviceTemplatesDir, "error", err)
	}

	notes := cfg.Notes
	if notes == nil {
		notes = make(map[string]string)
	}

	var pollers []*poller.Poller
	for _, sw := range cfg.ActiveSwitches() {
		p := startPolling(cache, sw.IP, sw.Name, sw.Model, cfg.RefreshInterval, logger, notes, histStore, cfg.DemoMode)
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
				logger.Warn("switch unreachable, live data unavailable", "ip", ip, "error", err)
				return
			}
			if sw.PSK != "" {
				if err := client.SetPSK(sw.PSK); err != nil {
					logger.Warn("invalid psk in config, continuing without encryption", "ip", ip, "error", err)
				} else {
					logger.Info("psk configured, write commands will use /enc", "ip", ip)
				}
			}

			info, err := client.ScrapeInformation()
			if err != nil {
				logger.Warn("initial scrape failed, waiting for next poll", "ip", ip, "error", err)
				return
			}
			logger.Info("connected, switching to live data", "ip", ip, "model", info.HWVer)

			p := startPollingWithClient(cache, client, ip, sw.Name, sw.Model, cfg.RefreshInterval, logger, notes, histStore, cfg.DemoMode)
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

func startPolling(cache *server.Cache, ip, name, model string, interval int, logger *slog.Logger, notes map[string]string, histStore *history.Store, demo bool) *poller.Poller {
	p := poller.NewWithNotes(cache, ip, name, model, interval, notes, logger)
	p.SetDemo(demo)
	if histStore != nil {
		p.SetHistoryStore(histStore)
	}
	p.Start()
	return p
}

func startPollingWithClient(cache *server.Cache, client *rtlplayground.Client, ip, name, model string, interval int, logger *slog.Logger, notes map[string]string, histStore *history.Store, demo bool) *poller.Poller {
	p := poller.NewWithClientAndNotes(cache, client, ip, name, model, interval, notes, logger)
	p.SetDemo(demo)
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
