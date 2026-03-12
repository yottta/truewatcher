package client

// Application represents the minimal set of fields that are needed to be decoded from the "app.query"
// TrueNAS method to be able to call the update of the Application.
type Application struct {
	Name             string `json:"name"`
	Id               string `json:"id"`
	UpgradeAvailable bool   `json:"upgrade_available"`
	LatestVersion    string `json:"latest_version"`
	Version          string `json:"version"`
}

// List is a generic list type for the TrueNAS query responses.
type List[T any] struct {
	Entries []T `json:"result"`
}
