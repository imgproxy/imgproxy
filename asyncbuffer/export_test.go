package asyncbuffer

// CondCh returns the internal channel of a Cond for testing.
func CondCh(c *Cond) chan struct{} {
	return c.ch
}

// LatchDone returns the internal done channel of a Latch for testing.
func LatchDone(l *Latch) chan struct{} {
	return l.done
}
