package main

import (
	"flag"
)

type ServerFlags struct {
	address string
}

func (flags *ServerFlags) Init() {
	flag.StringVar(&flags.address, "a", ":8080", "address and port to run server")

	flag.Parse()
}
