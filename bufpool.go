package main

import (
	"bytes"
	"runtime"
	"sort"
	"sync"
)

type intSlice []int

func (p intSlice) Len() int           { return len(p) }
func (p intSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p intSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type bufPool struct {
	name        string
	defaultSize int
	maxSize     int
	buffers     []*bytes.Buffer

	calls   intSlice
	callInd int

	mutex sync.Mutex
}

func newBufPool(name string, n int, defaultSize int) *bufPool {
	pool := bufPool{
		name:        name,
		defaultSize: defaultSize,
		buffers:     make([]*bytes.Buffer, n),
		calls:       make(intSlice, conf.BufferPoolCalibrationThreshold),
	}

	for i := range pool.buffers {
		pool.buffers[i] = new(bytes.Buffer)
	}

	return &pool
}

func (p *bufPool) calibrateAndClean() {
	sort.Sort(p.calls)

	pos := int(float64(len(p.calls)) * 0.95)
	score := p.calls[pos]

	p.callInd = 0
	p.maxSize = p.normalizeSize(score)

	p.defaultSize = maxInt(p.defaultSize, p.calls[0])
	p.maxSize = maxInt(p.defaultSize, p.maxSize)

	cleaned := false

	for i, buf := range p.buffers {
		if buf != nil && buf.Cap() > p.maxSize {
			p.buffers[i] = nil
			cleaned = true
		}
	}

	if cleaned {
		runtime.GC()
	}

	if prometheusEnabled {
		setPrometheusBufferDefaultSize(p.name, p.defaultSize)
		setPrometheusBufferMaxSize(p.name, p.maxSize)
	}
}

func (p *bufPool) Get(size int) *bytes.Buffer {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	size = p.normalizeSize(size)

	minSize, maxSize, minInd, maxInd := -1, -1, -1, -1

	for i := 0; i < len(p.buffers); i++ {
		if p.buffers[i] != nil {
			cap := p.buffers[i].Cap()

			if size > 0 && cap >= size && (minSize > cap || minSize == -1) {
				minSize = cap
				minInd = i
			}

			if cap > maxSize {
				maxSize = cap
				maxInd = i
			}
		}
	}

	var buf *bytes.Buffer

	switch {
	case minInd >= 0:
		// We found buffer with the desired size
		buf = p.buffers[minInd]
		p.buffers[minInd] = nil
	case maxInd >= 0:
		// We didn't find buffer with the desired size
		buf = p.buffers[maxInd]
		p.buffers[maxInd] = nil
	default:
		// We didn't find buffers at all
		buf = new(bytes.Buffer)
	}

	buf.Reset()

	growSize := maxInt(size, p.defaultSize)

	if growSize > buf.Cap() {
		buf.Grow(growSize)
	}

	return buf
}

func (p *bufPool) Put(buf *bytes.Buffer) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if buf.Len() > 0 {
		p.calls[p.callInd] = buf.Len()
		p.callInd++

		if p.callInd == len(p.calls) {
			p.calibrateAndClean()
		}
	}

	if p.maxSize > 0 && buf.Cap() > p.maxSize {
		return
	}

	for i, b := range p.buffers {
		if b == nil {
			p.buffers[i] = buf

			if prometheusEnabled && buf.Cap() > 0 {
				observePrometheusBufferSize(p.name, buf.Cap())
			}

			return
		}
	}
}

func (p *bufPool) normalizeSize(n int) int {
	return (n/bytes.MinRead + 2) * bytes.MinRead
}
