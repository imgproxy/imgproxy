package bufpool

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
)

var (
	testData     [][]byte
	testDataOnce sync.Once
	testMu       sync.Mutex
)

func initTestData() {
	testData = make([][]byte, 1000)
	for i := 6; i < 1000; i++ {
		testData[i] = make([]byte, i*1271)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(testData), func(i, j int) { testData[i], testData[j] = testData[j], testData[i] })
}

func BenchmarkBufpool(b *testing.B) {
	testMu.Lock()
	defer testMu.Unlock()

	config.Reset()

	testDataOnce.Do(initTestData)

	pool := New("test", 16, 0)

	b.ResetTimer()
	b.SetParallelism(16)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, bb := range testData {
				buf := pool.Get(len(bb), false)
				buf.Write(bb)
				pool.Put(buf)
			}
		}
	})
}
