package honeybadger

import "sync"

type contextSync struct {
	sync.RWMutex
	internal Context
}

func (context *contextSync) Update(other Context) {
	context.Lock()
	context.internal.Update(other)
	context.Unlock()
}

func newContextSync() *contextSync {
	instance := contextSync{
		internal: Context{},
	}

	return &instance
}
