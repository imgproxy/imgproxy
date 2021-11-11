package memory

import (
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/vips"
)

func LogStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Debugf(
		"GO MEMORY USAGE: Sys=%d HeapIdle=%d HeapInuse=%d",
		m.Sys/1024/1024, m.HeapIdle/1024/1024, m.HeapInuse/1024/1024,
	)

	log.Debugf(
		"VIPS MEMORY USAGE: Cur=%d Max=%d Allocs=%d",
		int(vips.GetMem())/1024/1024, int(vips.GetMemHighwater())/1024/1024, int(vips.GetAllocs()),
	)
}
