package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"hash"
	"sync"
)

type etagPool struct {
	mutex sync.Mutex
	top   *etagPoolEntry
}

type etagPoolEntry struct {
	hash hash.Hash
	enc  *json.Encoder
	next *etagPoolEntry
	b    []byte
}

func newEtagPool(n int) *etagPool {
	pool := new(etagPool)

	for i := 0; i < n; i++ {
		pool.grow()
	}

	return pool
}

func (p *etagPool) grow() {
	h := sha256.New()

	enc := json.NewEncoder(h)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")

	p.top = &etagPoolEntry{
		hash: h,
		enc:  enc,
		b:    make([]byte, 64),
		next: p.top,
	}
}

func (p *etagPool) Get() *etagPoolEntry {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.top == nil {
		p.grow()
	}

	entry := p.top
	p.top = p.top.next

	return entry
}

func (p *etagPool) Put(e *etagPoolEntry) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	e.next = p.top
	p.top = e
}

var eTagCalcPool *etagPool

func calcETag(ctx context.Context) ([]byte, context.CancelFunc) {
	c := eTagCalcPool.Get()
	cancel := func() { eTagCalcPool.Put(c) }

	c.hash.Reset()
	c.hash.Write(getImageData(ctx).Bytes())
	footprint := c.hash.Sum(nil)

	c.hash.Reset()
	c.hash.Write(footprint)
	c.hash.Write([]byte(version))
	c.enc.Encode(conf)
	c.enc.Encode(getProcessingOptions(ctx))

	hex.Encode(c.b, c.hash.Sum(nil))

	return c.b, cancel
}
