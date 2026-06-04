package defs

const (
	RequestsTotal     = "requests_total"
	RequestsTotalDesc = "A counter of the total number of HTTP requests imgproxy processed."

	StatusCodesTotal     = "status_codes_total"
	StatusCodesTotalDesc = "A counter of the response status codes."

	ErrorsTotal     = "errors_total"
	ErrorsTotalDesc = "A counter of the occurred errors separated by type."

	RequestDurationSeconds     = "request_duration_seconds"
	RequestDurationSecondsDesc = "A histogram of the response latency."

	RequestSpanDurationSeconds     = "request_span_duration_seconds"
	RequestSpanDurationSecondsDesc = "A histogram of the request spans duration separated by span name."

	Workers     = "workers"
	WorkersDesc = "A gauge of the number of running workers."

	RequestsInProgress     = "requests_in_progress"
	RequestsInProgressDesc = "A gauge of the number of requests currently being in progress."

	ImagesInProgress     = "images_in_progress"
	ImagesInProgressDesc = "A gauge of the number of images currently being in progress."

	WorkersUtilization     = "workers_utilization"
	WorkersUtilizationDesc = "A gauge of the workers utilization in percents."

	VipsMemoryBytes     = "vips_memory_bytes"
	VipsMemoryBytesDesc = "A gauge of the vips tracked memory usage in bytes."

	VipsMaxMemoryBytes     = "vips_max_memory_bytes"
	VipsMaxMemoryBytesDesc = "A gauge of the max vips tracked memory usage in bytes."

	VipsAllocs     = "vips_allocs"
	VipsAllocsDesc = "A gauge of the number of active vips allocations."

	ProcessResidentMemoryBytes     = "process_resident_memory_bytes"
	ProcessResidentMemoryBytesDesc = "Resident memory size in bytes."

	ProcessVirtualMemoryBytes     = "process_virtual_memory_bytes"
	ProcessVirtualMemoryBytesDesc = "Virtual memory size in bytes."

	GoMemstatsSysBytes     = "go_memstats_sys_bytes"
	GoMemstatsSysBytesDesc = "Number of bytes obtained from system."

	GoMemstatsHeapIdleBytes     = "go_memstats_heap_idle_bytes"
	GoMemstatsHeapIdleBytesDesc = "Number of heap bytes waiting to be used."

	GoMemstatsHeapInuseBytes     = "go_memstats_heap_inuse_bytes"
	GoMemstatsHeapInuseBytesDesc = "Number of heap bytes that are in use."

	GoGoroutines     = "go_goroutines"
	GoGoroutinesDesc = "Number of goroutines that currently exist."

	GoThreads     = "go_threads"
	GoThreadsDesc = "Number of OS threads created."
)
