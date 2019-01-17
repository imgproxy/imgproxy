package main

import (
	"bytes"
	"sync"
)

type bufPool struct {
	mutex sync.Mutex
	size  int
	top   *bufPoolEntry
}

type bufPoolEntry struct {
	buf  *bytes.Buffer
	next *bufPoolEntry
}

func newBufPool(n int, size int) *bufPool {
	pool := bufPool{size: size}

	for i := 0; i < n; i++ {
		pool.grow()
	}

	return &pool
}

func (p *bufPool) grow() {
	var buf *bytes.Buffer

	if p.size == 0 {
		buf = new(bytes.Buffer)
	} else {
		buf = bytes.NewBuffer(make([]byte, p.size, p.size))
	}

	p.top = &bufPoolEntry{buf: buf, next: p.top}
}

func (p *bufPool) get() *bytes.Buffer {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.top == nil {
		p.grow()
	}

	buf := p.top.buf
	buf.Reset()

	p.top = p.top.next

	return buf
}

func (p *bufPool) put(buf *bytes.Buffer) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.top = &bufPoolEntry{buf: buf, next: p.top}
}
