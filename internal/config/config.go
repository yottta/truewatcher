package config

import (
	"log/slog"
	"strings"
	"time"

	"github.com/yottta/go-core/env"
)

var defaultCheckDelay = 6 * time.Hour

type Config struct {
	URL         string
	Username    string
	Password    string
	APIKey      string
	CheckDelay  time.Duration
	Whitelisted []string
	Blacklisted []string
}

func LoadConfig() Config {
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
		slog.
			With(
				"given_value", rawCheckDelay,
				"default", defaultCheckDelay.String(),
				"error", err,
			).Warn("failed to parse the given value for CHECK_DELAY. Fallback to default")
		delay = defaultCheckDelay
	}
	ret.CheckDelay = delay
	// parse filters
	if wl := strings.TrimSpace(env.String("APP_WHITELIST")); wl != "" {
		vals := strings.Split(wl, ",")
		ret.Whitelisted = make([]string, 0, len(vals))
		for _, val := range vals {
			ret.Whitelisted = append(ret.Whitelisted, strings.TrimSpace(strings.ToLower(val)))
		}
	}
	if bl := strings.TrimSpace(env.String("APP_BLACKLIST")); bl != "" {
		vals := strings.Split(bl, ",")
		ret.Blacklisted = make([]string, 0, len(vals))
		for _, val := range vals {
			ret.Blacklisted = append(ret.Blacklisted, strings.TrimSpace(strings.ToLower(val)))
		}
	}

	return ret
}
