package main

import (
	"os"

	"github.com/brunostjohn/k8s-mc-loadbalancer/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	err := cmd.Execute()

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
