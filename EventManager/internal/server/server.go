package server

import (
	"log"

	"EventManager/internal/mqtt"
)

func Start() {
	klijent := mqtt.NoviKlijent()
	if klijent == nil {
		log.Fatal("Greška.")
	}

	token := klijent.Subscribe("topic/NoviPodaci", 2, nil)
	<- token.Done()
	if token.Error() != nil {
		log.Fatal("klijent.Subscribe(topic/NoviPodaci)", token.Error)
	}
}
