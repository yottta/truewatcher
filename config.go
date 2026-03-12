package main

import (
	"log/slog"
	"strings"
	"time"

	"github.com/yottta/go-core/env"
)

var defaultCheckDelay = 6 * time.Hour

type Config struct {
	URL        string
	Username   string
	Password   string
	APIKey     string
	CheckDelay time.Duration
	filtering  filter
}

func loadConfig() Config {
	ret := Config{
		URL:      env.String("TRUENAS_URL"),
		Username: env.String("TRUENAS_USERNAME"),
		Password: env.String("TRUENAS_PASSWORD"),
		APIKey:   env.String("TRUENAS_API_KEY"),
	}
	// parse the recheck delay
	rawCheckDelay := env.StringWithDefault("CHECK_DELAY", defaultCheckDelay.String())
	delay, err := time.ParseDuration(rawCheckDelay)
	if err != nil {
		slog.With("given_value", rawCheckDelay).With("default", defaultCheckDelay.String()).With("error", err).Warn("failed to parse the given value for CHECK_DELAY. Fallback to default")
		delay = defaultCheckDelay
	}
	ret.CheckDelay = delay
	// parse filters
	ret.filtering = filterChain{
		whitelistFilter(strings.Split(env.String("APP_WHITELIST"), ",")),
		blacklistFilter(strings.Split(env.String("APP_BLACKLIST"), ",")),
	}

	return ret
}
