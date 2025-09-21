package main

import (
	"os"
	"os/signal"
	"syscall"

	"EventManager/internal/server"
)

func main() {
	server.Start()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		<- sigs
		done <- true
	}()

	<- done
}
