package main

import (
	"os"
	"os/signal"

	"github.com/marksteve/telltheturtle/mailer"
)

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, os.Kill)
	mailer.Run(done)
}
