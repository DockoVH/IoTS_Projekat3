package server

import (
	"log"
	"net/http"
	"bytes"
	"encoding/json"
	"time"
	"io"

	emqx "github.com/eclipse/paho.mqtt.golang"
	"github.com/nats-io/nats.go"

	"Analytics/internal/mqtt"
)

type senzorPodatak struct {
	Temperatura float64
	Vlaznost float64
	Pm2_5 float64
	Pm10 float64
}

var (
	brojPodataka             = 0
	podaci       [][]float64 = [][]float64{
		[]float64{0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0, 0.0},
	}
)

func Start() {
	klijent := mqtt.NoviKlijent()
	if klijent == nil {
		return
	}
	natsKlijent, err := nats.Connect("nats:4222")
	if err != nil {
		log.Printf("nats.Connect() greška: %v\n", err)
		return
	}

	defer func() {
		natsKlijent.Close()
		token := klijent.Unsubscribe("topic/NoviPodaci")
		<-token.Done()
		if token.Error() != nil {
			log.Printf("klijent.Unsubscribe(\"topic/NoviPodaci\") greška: %v\n", token.Error())
		}
		klijent.Disconnect(250)
	}()

	ch := make(chan []byte, 1)

	token := klijent.Subscribe("topic/NoviPodaci", 1, func(client emqx.Client, msg emqx.Message) {
		log.Printf("mqtt: Primljen podatak: %v sa topica %s\n", string(msg.Payload()), msg.Topic())
		ch <- msg.Payload()
	})

	<-token.Done()
	if token.Error() != nil {
		log.Printf("Subscribe(\"topic/NoviPodaci\") greška: %v\n", token.Error())
	}

	url := "http://mlaas:9080/predict"

	for {
		select {
			case poruka, ok := <- ch:
				if !ok {
					break
				}

				var podatak senzorPodatak
				if err := json.Unmarshal(poruka, &podatak); err != nil {
					log.Printf("json.Unmarshal() greška: %v\n", err)
					continue
				}
				dodajPodatak(podatak)

				reqBody := struct {
					Podaci [][]float64
				} {
					Podaci: podaci,
				}

				jsonStr, err := json.Marshal(reqBody)
				if err != nil {
					log.Printf("json.Marshal(body) greška: %v\n", err)
					continue
				}

				req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
				if err != nil {
					log.Printf("http.NewRequest greška: %v\n", err)
					continue
				}

				req.Header.Set("Content-Type", "application/json")
				client := &http.Client{}
				client.Timeout = time.Second * 15

				resp, err := client.Do(req)
				if err != nil {
					log.Printf("client.Do() greška: %v\n", err)
					continue
				}

				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Printf("io.ReadAll(resp.Body) greška: %v\n", err)
					continue
				}
				log.Printf("[%v]: %s\n", resp.Status, string(body))
				natsKlijent.Publish("nats/Predikcija", body)
			default:
				continue
		}
	}
}

func dodajPodatak(sp senzorPodatak) {
	if brojPodataka < 10 {
		levo := podaci[0:brojPodataka]
		desno := [][]float64 { []float64 { sp.Temperatura, sp.Vlaznost, sp.Pm2_5, sp.Pm10 } }
		for i := brojPodataka; i < 9; i++ {
			desno = append(desno, []float64 { sp.Temperatura, sp.Vlaznost, sp.Pm2_5, sp.Pm10 })
		}

		podaci = append(levo, desno...)
		brojPodataka++

		return
	}

	podaci = append(podaci[1:], []float64 { sp.Temperatura, sp.Vlaznost, sp.Pm2_5, sp.Pm10 })
}
