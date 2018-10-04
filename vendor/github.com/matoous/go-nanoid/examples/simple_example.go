package main

import (
	"fmt"

	"github.com/matoous/go-nanoid"
)

func main() {
	id, err := gonanoid.Nanoid()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Generated id: %s\n", id)
}
