package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yottta/go-core/logging"
	"github.com/yottta/truewatcher/sdk"
)

// application represents the minimal set of fields that are needed to be decoded from the "app.query"
// TrueNAS method to be able to call the update of the application.
type application struct {
	Name             string `json:"name"`
	Id               string `json:"id"`
	UpgradeAvailable bool   `json:"upgrade_available"`
	LatestVersion    string `json:"latest_version"`
	Version          string `json:"version"`
}

// list is a generic list type for the TrueNAS query responses.
type list[T any] struct {
	Entries []T `json:"result"`
}

func main() {
	logging.Setup()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	cl, closer, err := connect()
	defer func() {
		closer() // done this way because 'closer' can be updated later during refreshing the connection
	}()
	if err != nil {
		slog.With("error", err).Error("failed on initial connection")
		os.Exit(1)
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Info("app watcher stopped")
				return
			case <-time.After(30 * time.Second):
				if !ping(cl) {
					// ensure that the old connection is cleaned up
					closer()
					// reconnect...
					cl, closer, err = connect()
					if err != nil {
						slog.With("error", err).Error("error reconnecting")
						continue
					}
					slog.Info("reconnected")
				}
			case <-time.After(6 * time.Hour):
				queryAndUpgrade(cl)
			}
		}
	}()

	defer stop()
	slog.Info("appwatcher started. ctrl+c to shut it down")
	<-ctx.Done()
	slog.Info("appwatcher stopped")
}

// connect uses the `TRUENAS_URL` environment variable to create a new websocket client.
func connect() (*sdk.Client, func(), error) {
	url := os.Getenv("TRUENAS_URL")
	closer := func() {}
	cl, err := sdk.NewClientWithCallback(url, false, func(i int64, i2 int64, m map[string]interface{}) {
		slog.With(
			"i", i,
			"i2", i2,
			"m", m,
		).Info("job received")
	})
	if err != nil {
		return nil, closer, err
	}
	closer = func() {
		_ = cl.Close()
	}
	if err := login(cl); err != nil {
		slog.With("error", err).Error("failed to login")
		return cl, closer, err
	}
	return cl, closer, nil
}

// login uses the given TrueNAS client and tries to login.
// To be able to properly authenticate, this needs one of the following combination set of environment variables:
//   - `TRUENAS_USERNAME` and `TRUENAS_PASSWORD`
//   - `TRUENAS_API_KEY`
func login(cl *sdk.Client) error {
	username := os.Getenv("TRUENAS_USERNAME")
	password := os.Getenv("TRUENAS_PASSWORD")
	apiKey := os.Getenv("TRUENAS_API_KEY")
	if err := cl.Login(username, password, apiKey); err != nil {
		return err
	}
	return cl.SubscribeToJobs()
}

// ping returns false if it failed to ping
func ping(cl *sdk.Client) bool {
	resp, err := cl.Ping()
	if err != nil {
		slog.With("error", err).Error("failed to ping")
		return false
	}
	slog.With("result", resp).Debug("ping result")
	return true
}

// queryAndUpgrade gets the applications that are reported with and upgrade available and calls
// upgrade on the TrueNAS client.
func queryAndUpgrade(cl *sdk.Client) {
	slog.Debug("looking for apps to update")
	apps, err := queryApps(cl)
	if err != nil {
		slog.With("error", err).Error("failed to query apps")
		return
	}
	slog.With("no_of_apps", len(apps.Entries)).Debug("apps returned")
	for _, app := range apps.Entries {
		slog.With(
			"app_name", app.Name,
			"app_id", app.Name,
			"upgrade_available", app.UpgradeAvailable,
		).Debug("app returned")
		ok, err := upgradeApp(cl, app)
		if err != nil {
			slog.With(
				"app_name", app.Name,
				"app_id", app.Name,
				"upgrade_available", app.UpgradeAvailable,
				"current_version", app.Version,
				"latest_version", app.LatestVersion,
				"error", err,
			).Error("error upgrading the app")
		}
		if ok {
			slog.With(
				"app_name", app.Name,
				"app_id", app.Name,
				"upgrade_available", app.UpgradeAvailable,
				"from_version", app.Version,
				"to_version", app.LatestVersion,
			).Info("app upgraded")
		}
	}
}

// queryApps gets only the applications with `upgrade_available=true` and returns an unmarshalled list.
func queryApps(cl *sdk.Client) (*list[application], error) {
	request := []any{
		[]any{
			[]any{"upgrade_available", "=", true},
		},
		map[string]any{
			"select": []string{"upgrade_available", "name", "id", "latest_version", "version"},
		},
	}
	resp, err := cl.Call("app.query", 10, request)
	if err != nil {
		return nil, err
	}
	var ret list[application]
	if err := json.Unmarshal(resp, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

// upgradeApp upgrades the application if there is an upgrade available for it.
func upgradeApp(cl *sdk.Client, app application) (bool, error) {
	if !app.UpgradeAvailable {
		return false, nil
	}
	request := []any{
		app.Name,
		map[string]any{
			"app_version": app.LatestVersion,
		},
	}
	resp, err := cl.Call("app.upgrade", 10, request)
	if err != nil {
		slog.With("error", err, "app", app.Name, "resp", resp).Error("failed to upgrade app")
		return false, err
	}
	return true, nil
}
