package main

import (
	"log"

	"SensorGenerator/internal/server"
)

func main() {
	if err := server.Healthcheck(); err != nil {
		log.Fatal("SensorGenerator nije uspešno pokrenut: ", err)
	}
	server.Pisi()
}
