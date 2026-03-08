package main

import "fmt"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func formattedVersion() string {
	return fmt.Sprintf("%s-%s from %s", version, commit, date)
}
