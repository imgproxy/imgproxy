package memory

import (
	"log/slog"
	"runtime"

	vipsstats "github.com/imgproxy/imgproxy/v3/vips/stats"
)

func LogStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	slog.Debug(
		"GO MEMORY USAGE",
		"sys", m.Sys/1024/1024,
		"heap_idle", m.HeapIdle/1024/1024,
		"heap_inuse", m.HeapInuse/1024/1024,
	)

	slog.Debug(
		"VIPS MEMORY USAGE",
		"cur", int(vipsstats.Memory())/1024/1024,
		"max", int(vipsstats.MemoryHighwater())/1024/1024,
		"allocs", int(vipsstats.Allocs()),
	)
}
