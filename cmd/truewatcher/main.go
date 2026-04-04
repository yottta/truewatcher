package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/yottta/go-core/logging"
	"github.com/yottta/truewatcher/internal/client"
	"github.com/yottta/truewatcher/internal/config"
)

func main() {
	logging.Setup()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	cfg := config.LoadConfig()
	cl := client.Client{
		URL:        cfg.URL,
		Username:   cfg.Username,
		Password:   cfg.Password,
		APIKey:     cfg.APIKey,
		CheckDelay: cfg.CheckDelay,
		Filtering: client.FilterChain{
			client.WhitelistFilter(cfg.Whitelisted),
			client.BlacklistFilter(cfg.Blacklisted),
		},
	}
	l := slog.With("version", formattedVersion())
	if len(cfg.Whitelisted) > 0 {
		l = l.With("whitelisted_apps", cfg.Whitelisted)
	}
	if len(cfg.Blacklisted) > 0 {
		l = l.With("blacklisted_apps", cfg.Blacklisted)
	}
	l = l.With("check_delay", cfg.CheckDelay.String())
	l.Info("started")
	if err := cl.MonitorApps(ctx); err != nil {
		l.With("error", err).Error("error monitoring the applications")
		os.Exit(1)
	}
	l.Info("stopped")
}
