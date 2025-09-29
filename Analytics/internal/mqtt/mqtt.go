package mqtt

import (
	"log"
	"time"

	emqx "github.com/eclipse/paho.mqtt.golang"
)

func NoviKlijent() emqx.Client {
        opts := emqx.NewClientOptions()
        opts.AddBroker("tcp://mqtt:1883")
        opts.SetClientID("iots_analytics")
        opts.SetKeepAlive(5 * time.Second)

        klijent := emqx.NewClient(opts)
        if token := klijent.Connect(); token.Wait() && token.Error() != nil {
                log.Print("NoviKlijent(): Greška: ", token.Error())
                return nil
        }

        return klijent
}
