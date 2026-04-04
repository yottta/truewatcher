package client

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/yottta/truewatcher/internal/client/sdk"
)

type Client struct {
	URL        string
	Username   string
	Password   string
	APIKey     string
	CheckDelay time.Duration
	Filtering  Filter
}

func (c *Client) MonitorApps(ctx context.Context) error {
	cl, closer, err := c.connect()
	defer func() {
		closer() // done this way because 'closer' can be updated later during refreshing the connection
	}()
	if err != nil {
		return err
	}
	t := time.NewTicker(c.CheckDelay)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); !errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		case <-time.After(30 * time.Second):
			if !c.ping(cl) {
				// ensure that the old connection is cleaned up
				closer()
				// reconnect...
				cl, closer, err = c.connect()
				if err != nil {
					slog.With("error", err).Error("error reconnecting")
					continue
				}
				slog.Info("reconnected")
			}
		case <-t.C:
			c.queryAndUpgrade(cl)
		}
	}
}

func (c *Client) connect() (*sdk.Client, func(), error) {
	closer := func() {}
	cl, err := sdk.NewClient(c.URL, false)
	if err != nil {
		return nil, closer, err
	}
	closer = func() {
		_ = cl.Close()
	}
	if err := c.login(cl); err != nil {
		slog.With("error", err).Error("failed to login")
		return cl, closer, err
	}
	return cl, closer, nil
}

// login uses the given TrueNAS client.
func (c *Client) login(cl *sdk.Client) error {
	if err := cl.Login(c.Username, c.Password, c.APIKey); err != nil {
		return err
	}
	return cl.SubscribeToJobs()
}

// ping returns false if it failed to ping
func (c *Client) ping(cl *sdk.Client) bool {
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
func (c *Client) queryAndUpgrade(cl *sdk.Client) {
	slog.Debug("looking for apps to update")
	apps, err := c.queryApps(cl)
	if err != nil {
		slog.With("error", err).Error("failed to query apps")
		return
	}
	slog.With("no_of_apps", len(apps.Entries)).Debug("apps returned")
	for _, app := range apps.Entries {
		slog.With(
			"app_name", app.Name,
			"app_id", app.Id,
			"upgrade_available", app.UpgradeAvailable,
			"current_version", app.Version,
			"latest_version", app.LatestVersion,
		).Debug("app returned")
		ok, err := c.upgradeApp(cl, app)
		if err != nil {
			slog.With(
				"app_name", app.Name,
				"app_id", app.Name,
				"current_version", app.Version,
				"latest_version", app.LatestVersion,
				"error", err,
			).Error("error upgrading the app")
		}
		if ok {
			slog.With(
				"app_name", app.Name,
				"app_id", app.Name,
				"from_version", app.Version,
				"to_version", app.LatestVersion,
			).Info("app upgraded")
		}
	}
}

// queryApps gets only the applications with `upgrade_available=true` and returns an unmarshalled list.
func (c *Client) queryApps(cl *sdk.Client) (*List[Application], error) {
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
	var ret List[Application]
	if err := json.Unmarshal(resp, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

// upgradeApp upgrades the application if there is an upgrade available for it.
func (c *Client) upgradeApp(cl *sdk.Client, app Application) (bool, error) {
	if !app.UpgradeAvailable {
		return false, nil
	}
	if !c.Filtering.Allowed(app) {
		slog.With(
			"app_name", app.Name,
			"app_id", app.Id,
			"upgrade_available", app.UpgradeAvailable,
		).Info("app filtered out")
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
