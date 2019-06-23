package main

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"sync"
)

type gzipPool struct {
	mutex sync.Mutex
	top   *gzipPoolEntry
}

type gzipPoolEntry struct {
	gz   *gzip.Writer
	next *gzipPoolEntry
}

func newGzipPool(n int) *gzipPool {
	pool := new(gzipPool)

	for i := 0; i < n; i++ {
		pool.grow()
	}

	return pool
}

func (p *gzipPool) grow() {
	gz, err := gzip.NewWriterLevel(ioutil.Discard, conf.GZipCompression)
	if err != nil {
		logFatal("Can't init GZip compression: %s", err)
	}

	p.top = &gzipPoolEntry{
		gz:   gz,
		next: p.top,
	}
}

func (p *gzipPool) Get(w io.Writer) *gzip.Writer {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.top == nil {
		p.grow()
	}

	gz := p.top.gz
	gz.Reset(w)

	p.top = p.top.next

	return gz
}

func (p *gzipPool) Put(gz *gzip.Writer) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	gz.Reset(ioutil.Discard)

	p.top = &gzipPoolEntry{gz: gz, next: p.top}
}
