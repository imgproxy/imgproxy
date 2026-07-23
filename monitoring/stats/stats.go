package stats

import (
	"sync/atomic"
)

// Stats holds statistics counters thread safely
type Stats struct {
	requestsInProgress atomic.Int64
	imagesInProgress   atomic.Int64
	WorkersNumber      int
}

// New creates a new Stats instance
func New(workersNumber int) *Stats {
	return &Stats{
		WorkersNumber: workersNumber,
	}
}

// RequestsInProgress returns the current number of requests in progress
func (s *Stats) RequestsInProgress() float64 {
	return float64(s.requestsInProgress.Load())
}

// IncRequestsInProgress increments the requests in progress counter
func (s *Stats) IncRequestsInProgress() {
	s.requestsInProgress.Add(1)
}

// DecRequestsInProgress decrements the requests in progress counter
func (s *Stats) DecRequestsInProgress() {
	s.requestsInProgress.Add(-1)
}

// ImagesInProgress returns the current number of images being processed
func (s *Stats) ImagesInProgress() float64 {
	return float64(s.imagesInProgress.Load())
}

// IncImagesInProgress increments the images in progress counter
func (s *Stats) IncImagesInProgress() {
	s.imagesInProgress.Add(1)
}

// DecImagesInProgress decrements the images in progress counter
func (s *Stats) DecImagesInProgress() {
	s.imagesInProgress.Add(-1)
}

// WorkersUtilization returns the current workers utilization percentage
func (s *Stats) WorkersUtilization() float64 {
	if s.WorkersNumber == 0 {
		return 0.0
	}
	return s.RequestsInProgress() / float64(s.WorkersNumber) * 100.0
}
