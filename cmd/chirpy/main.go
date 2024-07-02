package main

import (
	"flag"

	"github.com/Quorum-Code/chirpy/internal/webserver"
)

type RunConfig struct {
	Debug bool
}

func main() {
	rcfg := RunConfig{}

	rcfg.Debug = *flag.Bool("debug", false, "Enable debug mode")

	webserver.StartServer(rcfg.Debug)
}
