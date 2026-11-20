package errorreport

import (
	"net/http"
	"sync"

	"github.com/imgproxy/imgproxy/v4/errctx"
)

// TestReport is one captured call to Reporter.Report.
type TestReport struct {
	Err  errctx.Error
	Req  *http.Request
	Meta map[string]any
}

// TestReporter captures what a Reporter built via NewTestReporter reports. Concurrency-safe.
type TestReporter struct {
	mu      sync.Mutex
	reports []TestReport
}

func (t *TestReporter) Report(err errctx.Error, req *http.Request, meta map[string]any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.reports = append(t.reports, TestReport{Err: err, Req: req, Meta: meta})
}

func (t *TestReporter) Close() {}

// Reports returns all captured reports.
func (t *TestReporter) Reports() []TestReport {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]TestReport(nil), t.reports...)
}

// Called reports whether Report was invoked at least once.
func (t *TestReporter) Called() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.reports) > 0
}

// NewTestReporter returns a Reporter wired to a TestReporter, plus the TestReporter itself.
func NewTestReporter() (*Reporter, *TestReporter) {
	tr := &TestReporter{}
	return &Reporter{reporters: []reporter{tr}}, tr
}
