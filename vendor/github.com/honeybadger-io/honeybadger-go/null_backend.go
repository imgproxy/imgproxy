package honeybadger

// nullBackend implements the Backend interface but swallows errors and does not
// send them to Honeybadger.
type nullBackend struct{}

// Ensure nullBackend implements Backend.
var _ Backend = &nullBackend{}

// NewNullBackend creates a backend which swallows all errors and does not send
// them to Honeybadger. This is useful for development and testing to disable
// sending unnecessary errors.
func NewNullBackend() Backend {
	return nullBackend{}
}

// Notify swallows error reports, does nothing, and returns no error.
func (b nullBackend) Notify(_ Feature, _ Payload) error {
	return nil
}
