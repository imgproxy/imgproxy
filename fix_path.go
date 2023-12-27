package main

import (
	"fmt"
	"regexp"
	"strings"
)

var fixPathRe = regexp.MustCompile(`/plain/(\S+)\:/([^/])`)

func fixPath(path string) string {
	for _, match := range fixPathRe.FindAllStringSubmatch(path, -1) {
		repl := fmt.Sprintf("/plain/%s://", match[1])
		if match[1] == "local" {
			repl += "/"
		}
		repl += match[2]
		path = strings.Replace(path, match[0], repl, 1)
	}

	return path
}
