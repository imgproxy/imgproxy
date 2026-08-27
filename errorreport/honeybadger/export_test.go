package honeybadger

import (
	"sync"

	honeybadgervendor "github.com/honeybadger-io/honeybadger-go"
)

// ReporterIface is the internal reporter type, exported for testing.
type ReporterIface = reporter

// NotifyCapturingBackend extends honeybadgervendor.TestBackend to also
// capture notices sent through the classic Notify path, which TestBackend
// itself ignores.
type NotifyCapturingBackend struct {
	honeybadgervendor.TestBackend

	mu      sync.Mutex
	notices []*honeybadgervendor.Notice
}

func (b *NotifyCapturingBackend) Notify(feature honeybadgervendor.Feature, payload honeybadgervendor.Payload) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if notice, ok := payload.(*honeybadgervendor.Notice); ok {
		b.notices = append(b.notices, notice)
	}

	return b.TestBackend.Notify(feature, payload)
}

// LastNotice returns the most recently captured notice, or nil if none was
// captured yet.
func (b *NotifyCapturingBackend) LastNotice() *honeybadgervendor.Notice {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.notices) == 0 {
		return nil
	}

	return b.notices[len(b.notices)-1]
}
