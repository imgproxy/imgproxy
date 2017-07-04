package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	s := &http.Server{
		Addr:           conf.Bind,
		Handler:        newHttpHandler(),
		ReadTimeout:    time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(conf.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Starting server at %s\n", conf.Bind)

	log.Fatal(s.ListenAndServe())
}
