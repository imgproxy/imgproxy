package stats

import "sync/atomic"

var (
	requestsInProgress int64
	imagesInProgress   int64
)

func RequestsInProgress() float64 {
	return float64(atomic.LoadInt64(&requestsInProgress))
}

func IncRequestsInProgress() {
	atomic.AddInt64(&requestsInProgress, 1)
}

func DecRequestsInProgress() {
	atomic.AddInt64(&requestsInProgress, -1)
}

func ImagesInProgress() float64 {
	return float64(atomic.LoadInt64(&imagesInProgress))
}

func IncImagesInProgress() {
	atomic.AddInt64(&imagesInProgress, 1)
}

func DecImagesInProgress() {
	atomic.AddInt64(&imagesInProgress, -1)
}
