package honeybadger

type envelope func() error

type worker interface {
	Push(envelope) error
	Flush()
}
