package defs

const (
	RequestsTotal              = "requests_total"
	StatusCodesTotal           = "status_codes_total"
	ErrorsTotal                = "errors_total"
	RequestDurationSeconds     = "request_duration_seconds"
	RequestSpanDurationSeconds = "request_span_duration_seconds"
	Workers                    = "workers"
	RequestsInProgress         = "requests_in_progress"
	ImagesInProgress           = "images_in_progress"
	WorkersUtilization         = "workers_utilization"
	VipsMemoryBytes            = "vips_memory_bytes"
	VipsMaxMemoryBytes         = "vips_max_memory_bytes"
	VipsAllocs                 = "vips_allocs"
	ProcessResidentMemoryBytes = "process_resident_memory_bytes"
	ProcessVirtualMemoryBytes  = "process_virtual_memory_bytes"
	GoMemstatsSysBytes         = "go_memstats_sys_bytes"
	GoMemstatsHeapIdleBytes    = "go_memstats_heap_idle_bytes"
	GoMemstatsHeapInuseBytes   = "go_memstats_heap_inuse_bytes"
	GoGoroutines               = "go_goroutines"
	GoThreads                  = "go_threads"
)
