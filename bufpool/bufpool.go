package bufpool

// Based on https://github.com/valyala/bytebufferpool ideas

import (
	"bytes"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/metrics"
)

const (
	minBitSize = 6 // 2**6=64 is min bytes.Buffer capacity
	steps      = 20

	minSize = 1 << minBitSize
)

var entriesPool = sync.Pool{
	New: func() any {
		return new(entry)
	},
}

type entry struct {
	buf        *bytes.Buffer
	prev, next *entry
}

type Pool struct {
	name        string
	defaultSize int
	maxSize     uint64
	root        *entry

	maxLen int

	calls    [steps]uint64
	tmpCalls [steps]uint64
	callsNum uint64

	storeMu       sync.Mutex
	calibratingMu sync.Mutex
}

func New(name string, n int, defaultSize int) *Pool {
	pool := Pool{
		name:        name,
		defaultSize: defaultSize,
		root:        &entry{},
		maxLen:      n,
	}

	return &pool
}

func (p *Pool) insert(buf *bytes.Buffer) {
	e := entriesPool.Get().(*entry)
	e.buf = buf
	e.next = p.root.next
	e.prev = p.root

	p.root.next = e
}

func (p *Pool) remove(e *entry) {
	if e.next != nil {
		e.next.prev = e.prev
	}

	e.prev.next = e.next

	saveEntry(e)
}

func (p *Pool) calibrateAndClean() {
	if !p.calibratingMu.TryLock() {
		return
	}
	defer p.calibratingMu.Unlock()

	var callsSum uint64
	for i := 0; i < steps; i++ {
		calls := atomic.SwapUint64(&p.calls[i], 0)
		callsSum += calls
		p.tmpCalls[i] = calls
	}

	if callsSum < uint64(config.BufferPoolCalibrationThreshold) {
		return
	}

	atomic.StoreUint64(&p.callsNum, 0)

	defSum := uint64(float64(callsSum) * 0.5)
	maxSum := uint64(float64(callsSum) * 0.95)

	defStep := -1
	maxStep := -1

	callsSum = 0
	for i := 0; i < steps; i++ {
		callsSum += p.tmpCalls[i]

		if defStep < 0 && callsSum > defSum {
			defStep = i
		}

		if callsSum > maxSum {
			maxStep = i
			break
		}
	}

	p.defaultSize = minSize << defStep
	p.maxSize = minSize << maxStep

	maxSize := int(p.maxSize)

	metrics.SetBufferDefaultSize(p.name, p.defaultSize)
	metrics.SetBufferMaxSize(p.name, maxSize)

	p.storeMu.Lock()
	storeUnlocked := false
	defer func() {
		if !storeUnlocked {
			p.storeMu.Unlock()
		}
	}()

	cleaned := false
	last := p.root

	poolLen := 0

	for entry := p.root.next; entry != nil; entry = last.next {
		if poolLen >= p.maxLen || entry.buf.Cap() > maxSize {
			last.next = entry.next
			saveEntry(entry)

			cleaned = true
		} else {
			last.next = entry
			entry.prev = last
			last = entry

			poolLen++
		}
	}

	// early unlock
	p.storeMu.Unlock()
	storeUnlocked = true

	if cleaned {
		runtime.GC()
	}
}

func (p *Pool) Get(size int, grow bool) *bytes.Buffer {
	p.storeMu.Lock()
	storeUnlocked := false
	defer func() {
		if !storeUnlocked {
			p.storeMu.Unlock()
		}
	}()

	best := (*entry)(nil)
	bestCap := -1

	min := (*entry)(nil)
	minCap := -1

	for entry := p.root.next; entry != nil; entry = entry.next {
		cap := entry.buf.Cap()

		if size > 0 {
			// If we know the required size, pick a buffer with the smallest size
			// that is larger than the requested size
			if cap >= size && (bestCap > cap || best == nil) {
				best = entry
				bestCap = cap
			}

			if cap < minCap || minCap == -1 {
				min = entry
				minCap = cap
			}
		} else if cap > bestCap {
			// If we don't know the requested size, pick a largest buffer
			best = entry
			bestCap = cap
		}
	}

	var buf *bytes.Buffer

	switch {
	case best != nil:
		buf = best.buf
		p.remove(best)
	case min != nil:
		buf = min.buf
		p.remove(min)
	default:
		buf = new(bytes.Buffer)
	}

	// early unlock
	p.storeMu.Unlock()
	storeUnlocked = true

	buf.Reset()

	growSize := p.defaultSize
	if grow {
		growSize = imath.Max(p.normalizeCap(size), growSize)
	}

	// Grow the buffer only if we know the requested size and it is smaller than
	// or equal to the grow size. Otherwise we'll grow the buffer twice
	if size > 0 && size <= growSize && growSize > buf.Cap() {
		buf.Grow(growSize)
	}

	return buf
}

func (p *Pool) Put(buf *bytes.Buffer) {
	bufLen := buf.Len()
	bufCap := buf.Cap()

	if bufLen > 0 {
		ind := index(bufLen)

		atomic.AddUint64(&p.calls[ind], 1)

		if atomic.AddUint64(&p.callsNum, 1) >= uint64(config.BufferPoolCalibrationThreshold) {
			p.calibrateAndClean()
		}
	}

	size := buf.Cap()
	maxSize := int(atomic.LoadUint64(&p.maxSize))
	if maxSize > 0 && size > maxSize {
		return
	}

	if bufLen > 0 {
		metrics.ObserveBufferSize(p.name, bufCap)
	}

	p.storeMu.Lock()
	defer p.storeMu.Unlock()

	p.insert(buf)
}

// GrowBuffer growth capacity of the buffer to the normalized provided value
func (p *Pool) GrowBuffer(buf *bytes.Buffer, cap int) {
	cap = p.normalizeCap(cap)
	if buf.Cap() < cap {
		buf.Grow(cap - buf.Len())
	}
}

func (p *Pool) normalizeCap(cap int) int {
	// Don't normalize cap if it's larger than maxSize
	// since we'll throw this buf out anyway
	maxSize := int(atomic.LoadUint64(&p.maxSize))
	if maxSize > 0 && cap > maxSize {
		return cap
	}

	ind := index(cap)
	return imath.Max(cap, minSize<<ind)
}

func saveEntry(e *entry) {
	e.buf = nil
	e.next = nil
	e.prev = nil
	entriesPool.Put(e)
}

func index(n int) int {
	n--
	n >>= minBitSize
	idx := 0
	for n > 0 {
		n >>= 1
		idx++
	}
	if idx >= steps {
		idx = steps - 1
	}
	return idx
}
