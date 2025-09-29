package server

import (
	"fmt"
	"time"
	"net/http"
	"strings"
	"log"

	"github.com/gorilla/websocket"
	emqx "github.com/eclipse/paho.mqtt.golang"
	"github.com/nats-io/nats.go"

	"MqttClient/internal/mqtt"
)

type senzorPodatak struct {
	Id int32
	Vreme time.Time
	Temperatura float32
	VlaznostVazduha float32
	Pm2_5 float32
	Pm10 float32
}

const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	upg = websocket.Upgrader {
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type Server struct {
	port int
}

func NoviServer(port int) *http.Server {
	noviServer := Server {
		port: port,
	}

	server := &http.Server {
		Addr: fmt.Sprintf("0.0.0.0:%d", noviServer.port),
		Handler: noviServer.HandlerInit(),
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return server
}

func (s *Server) HandlerInit() http.Handler {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))

	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/static")
		r.URL.Path = path
		fs.ServeHTTP(w, r)
	})

	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/ws", handleWs)

	return mux
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	log.Print(r.URL)

	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "index.html")
}

func handleWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upg.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("handleWs() upgrader: ", err)
	}

	ch := make(chan []byte, 4)

	klijentTemp := subscribe(ch, "topic/IznadGranice/Temperatura", '0')
	klijentVlaznost := subscribe(ch, "topic/IznadGranice/VlaznostVazduha", '1')
	klijentPm2_5 := subscribe(ch, "topic/IznadGranice/Pm2_5", '2')
	klijentPm10 := subscribe(ch, "topic/IznadGranice/Pm10", '3')
	if klijentTemp == nil || klijentVlaznost == nil || klijentPm2_5 == nil || klijentPm10 == nil {
		return
	}

	natsKlijent, natsSub := natsSubscribe(ch, "nats/Predikcija", '4')
	if natsKlijent == nil || natsSub == nil {
		return
	}

	defer func() {
		go unsubscribe(klijentTemp, "topic/IznadGranice/Temperatura")
		go unsubscribe(klijentVlaznost, "topic/IznadGranice/VlaznostVazduha")
		go unsubscribe(klijentPm2_5, "topic/IznadGranice/Pm2_5")
		go unsubscribe(klijentPm10, "topic/IznadGranice/Pm10")

		natsSub.Unsubscribe()
		natsKlijent.Drain()
	}()

	for {
		select {
			case poruka := <-ch:
				conn.SetWriteDeadline(time.Now().Add(writeWait))

				w, err := conn.NextWriter(websocket.BinaryMessage)
				if err != nil {
					log.Print("conn.NextWriter() greška: ", err)
					return
				}

				w.Write(poruka)
				if err = w.Close(); err != nil {
					log.Print("w.Close() greška: ", err)
				}
			default:
				continue
		}
	}
}

func subscribe(ch chan []byte, topic string, oznaka byte) emqx.Client {
	klijent := mqtt.NoviKlijent(oznaka)
	if klijent != nil {
		token := klijent.Subscribe(topic, 1, func (client emqx.Client, msg emqx.Message) {
			log.Printf("Primljen podatak: \n\tTopic: %v\n\tPayload: %v\n", msg.Topic(), string(msg.Payload()))
			rezultat := append([]byte{ oznaka } , msg.Payload()...)

			ch <- rezultat
		})

		go func() {
			<-token.Done()
			if token.Error() != nil {
				log.Printf("Subscribe(%s) greška: %v\n", topic, token.Error())
			}
		}()
	}

	return klijent
}

func unsubscribe(klijent emqx.Client, topic string) {
	token := klijent.Unsubscribe(topic)

	<-token.Done()
	if token.Error() != nil {
		log.Printf("Unsubscribe(\"%s\") greška: %v\n", topic, token.Error())
	}
	klijent.Disconnect(250)
}

func natsSubscribe(ch chan[]byte, topic string, oznaka byte) (*nats.Conn, *nats.Subscription) {
	natsKlijent, err := nats.Connect("nats:4222")

	sub, err := natsKlijent.Subscribe(topic, func(msg *nats.Msg) {
		log.Printf("Primljen podatak: %v, sa topic-a: %v\n", string(msg.Data), msg.Subject)

		rezultat := append([]byte{ oznaka }, msg.Data...)
		ch <- rezultat
	})
	if err != nil {
		log.Printf("nats.SubscribeSync(\"%s\") greška: %v\n", topic, err)
		return nil, nil
	}

	return natsKlijent, sub
}
