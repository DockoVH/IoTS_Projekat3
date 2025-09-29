package main

import (
	"os"
	"os/signal"
	"syscall"

	"Analytics/internal/server"
)

func main() {
	go server.Start()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		<- sigs
		done <- true
	}()

	<- done
}
