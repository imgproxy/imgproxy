package honeybadger

// Context is used to send extra data to Honeybadger.
type Context hash

// Update applies the values in other Context to context.
func (context Context) Update(other Context) {
	for k, v := range other {
		context[k] = v
	}
}
