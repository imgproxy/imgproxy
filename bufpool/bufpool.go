package bufpool

import (
	"bytes"
	"runtime"
	"sort"
	"sync"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/metrics/prometheus"
)

type intSlice []int

func (p intSlice) Len() int           { return len(p) }
func (p intSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p intSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type Pool struct {
	name        string
	defaultSize int
	maxSize     int
	buffers     []*bytes.Buffer

	calls   intSlice
	callInd int

	mutex sync.Mutex
}

func New(name string, n int, defaultSize int) *Pool {
	pool := Pool{
		name:        name,
		defaultSize: defaultSize,
		buffers:     make([]*bytes.Buffer, n),
		calls:       make(intSlice, config.BufferPoolCalibrationThreshold),
	}

	for i := range pool.buffers {
		pool.buffers[i] = new(bytes.Buffer)
	}

	return &pool
}

func (p *Pool) calibrateAndClean() {
	sort.Sort(p.calls)

	pos := int(float64(len(p.calls)) * 0.95)
	score := p.calls[pos]

	p.callInd = 0
	p.maxSize = p.normalizeSize(score)

	p.defaultSize = imath.Max(p.defaultSize, p.calls[0])
	p.maxSize = imath.Max(p.defaultSize, p.maxSize)

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

	prometheus.SetBufferDefaultSize(p.name, p.defaultSize)
	prometheus.SetBufferMaxSize(p.name, p.maxSize)
}

func (p *Pool) Get(size int) *bytes.Buffer {
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

	growSize := imath.Max(size, p.defaultSize)

	if growSize > buf.Cap() {
		buf.Grow(growSize)
	}

	return buf
}

func (p *Pool) Put(buf *bytes.Buffer) {
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

			if buf.Cap() > 0 {
				prometheus.ObserveBufferSize(p.name, buf.Cap())
			}

			return
		}
	}
}

func (p *Pool) normalizeSize(n int) int {
	return (n/bytes.MinRead + 2) * bytes.MinRead
}
