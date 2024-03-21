package main

import (
	api "github.com/hotnops/apeman/src/api/src"
)

func main() {
	s := new(api.Server)
	s.InitializeServer()
	s.Start()
}
