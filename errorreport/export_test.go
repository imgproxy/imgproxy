package errorreport

// ReporterIface is the internal reporter interface, exported for testing.
type ReporterIface = reporter

// NewWithReporters builds a Reporter with the given reporters, for testing.
func NewWithReporters(reporters ...ReporterIface) *Reporter {
	return &Reporter{reporters: reporters}
}
