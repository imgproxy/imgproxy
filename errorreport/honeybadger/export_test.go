package honeybadger

import "github.com/honeybadger-io/honeybadger-go"

// ReporterIface is the internal reporter type, exported for testing.
type ReporterIface = reporter

// NewWithClient builds a reporter around an already-configured Honeybadger client, for testing.
func NewWithClient(client *honeybadger.Client) *reporter {
	return &reporter{client: client}
}
