package main

import (
	api "github.com/hotnops/apeman/go/internal/api"
)

func main() {
	s := new(api.Server)
	s.InitializeServer()
	s.Start()
}
