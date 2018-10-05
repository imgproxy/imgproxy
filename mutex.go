package main

type mutex chan struct{}

func newMutex(size int) mutex {
	return make(mutex, size)
}

func (m mutex) Lock() {
	m <- struct{}{}
}

func (m mutex) Unock() {
	<-m
}
