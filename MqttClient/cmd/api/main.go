package main

import (
	"log"

	"MqttClient/internal/server"
)

func main() {
	port := 8085
	server := server.NoviServer(port)
	log.Print("Server pokrenut na adresi: ", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("ListenAndServe() greška: ", err)
	}
}
