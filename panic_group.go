package main

import "sync"

type panicGroup struct {
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

func (g *panicGroup) Wait() error {
	g.wg.Wait()
	return g.err
}

func (g *panicGroup) Go(f func() error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					g.errOnce.Do(func() {
						g.err = err
					})
				} else {
					panic(r)
				}
			}
		}()

		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
			})
		}
	}()
}
