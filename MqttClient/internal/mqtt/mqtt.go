package mqtt

import (
	"fmt"
	"log"
	"time"

	emqx "github.com/eclipse/paho.mqtt.golang"
)

func NoviKlijent(id byte) emqx.Client {
	opts := emqx.NewClientOptions()
	opts.AddBroker("tcp://mqtt:1883")
	opts.SetClientID(fmt.Sprintf("iots_mqtt_client_%v", id))
	opts.SetKeepAlive(5 * time.Second)

	klijent := emqx.NewClient(opts)
	if token := klijent.Connect(); token.Wait() && token.Error() != nil {
		log.Print("NoviKlijent() greška: ", token.Error())
		return nil
	}

	return klijent
}
