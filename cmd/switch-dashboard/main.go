package main

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/byte4geek/switch-dashboard/internal/config"
	"github.com/byte4geek/switch-dashboard/internal/poller"
	"github.com/byte4geek/switch-dashboard/internal/rtlplayground"
	"github.com/byte4geek/switch-dashboard/internal/server"
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	al := &appLogger{logger: logger}

	cfg, err := config.Load("config.json")
	if err != nil {
		logger.Warn("cannot load config", "error", err)
		cfg = &config.Config{}
		cfg.SetDefaults()
	}

	cache := server.NewCache()
	cfgProvider := &appConfig{
		title:          cfg.Title,
		refresh:        cfg.RefreshInterval,
		version:        "0.1.0",
	}

	srv := server.NewServer(cache, cfgProvider, al, nil, mustStaticFS())

	// Start poller for each active switch
	for _, sw := range cfg.ActiveSwitches() {
		sw := sw
		go func() {
			ip := sw.IP
			password := sw.Password
			logger.Info("connecting to switch", "ip", ip)

			client, err := rtlplayground.New(ip, password)
			if err != nil {
				logger.Error("failed to connect", "ip", ip, "error", err)
				startPolling(cache, ip, sw.Name, sw.Model, cfg.RefreshInterval, logger)
				return
			}

			// Initial scrape
			info, err := client.ScrapeInformation()
			if err != nil {
				logger.Error("initial scrape failed", "ip", ip, "error", err)
				startPolling(cache, ip, sw.Name, sw.Model, cfg.RefreshInterval, logger)
				return
			}
			logger.Info("connected to switch", "ip", ip, "model", info.HWVer, "hostname", info.Hostname)

			startPollingWithClient(cache, client, ip, sw.Name, sw.Model, cfg.RefreshInterval, logger)
		}()
	}

	addr := ":8080"
	logger.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, srv.Router); err != nil {
		logger.Error("server failed", "error", err)
	}
}

func startPolling(cache *server.Cache, ip, name, model string, interval int, logger *slog.Logger) {
	p := poller.New(cache, ip, name, model, interval)
	p.Start()
}

func startPollingWithClient(cache *server.Cache, client *rtlplayground.Client, ip, name, model string, interval int, logger *slog.Logger) {
	p := poller.NewWithClient(cache, client, ip, name, model, interval)
	p.Start()
}

func mustStaticFS() fs.FS {
	dir := "./static"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return os.DirFS(".")
}
