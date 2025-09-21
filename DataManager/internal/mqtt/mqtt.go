package mqtt

import (
	"log"
	"time"

	emqx "github.com/eclipse/paho.mqtt.golang"
)

var messagePubHandler emqx.MessageHandler = func(client emqx.Client, msg emqx.Message) {
	log.Printf("Primljena poruka: %s sa topic-a: %s\n", msg.Payload(), msg.Topic())
}

func NoviKlijent() emqx.Client {
	opts := emqx.NewClientOptions()
	opts.AddBroker("tcp://mqtt:1883")
	opts.SetClientID("iots_data_manager")
	opts.SetKeepAlive(5 * time.Second)

	opts.SetDefaultPublishHandler(messagePubHandler)

	klijent := emqx.NewClient(opts)
	if token := klijent.Connect(); token.Wait() && token.Error() != nil {
		log.Print("NoviKlijent(): Greška: ", token.Error())
		return nil
	}

	return klijent
}
