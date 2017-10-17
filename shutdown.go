// +build go1.8

package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

func shutdownServer(s *http.Server) {
	log.Println("Shutting down the server...")

	ctx, close := context.WithTimeout(context.Background(), 5*time.Second)
	defer close()

	s.Shutdown(ctx)
}
