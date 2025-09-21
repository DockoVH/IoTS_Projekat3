package server

import (
	"fmt"
	"os"
	"bufio"
	"log"
	"time"
	"strings"
	"strconv"
	"math"
	"encoding/json"
	"net/http"
	"bytes"
	"io"
	"errors"
)

type senzorPodatak struct {
	Temperatura float64
	VlaznostVazduha float64
	Pm25 float64
	Pm10 float64
}

func Healthcheck() error {
	greska := 0
	for {
		if greska == 10 {
			return errors.New(fmt.Sprintf("Konektovanje sa Gateway servisom nije uspelo nakon %v pokušaja.", greska))
		}

		res, err := http.Get("http://gateway:8080/api/SenzorPodaci/VratiSenzorPodatak/0")
		if err != nil {
			log.Print("[Healthcheck]: Greška: ", err)
			greska++

			time.Sleep(10 * time.Second)
			continue
		}

		if res.StatusCode == 200 {
			return nil
		}

		greska++
		time.Sleep(10 * time.Second)
	}
}

func Pisi() {
	for i := 1; i <= 50; i++ {
		putanja := fmt.Sprintf("/api/cmd/api/datasetovi/%v.csv", i)
		log.Print(putanja)

		url := "http://gateway:8080/api/SenzorPodaci/DodajSenzorPodatak"

		f, err := os.Open(putanja)
		if err != nil {
			log.Printf("os.Open(%v) greška: %v\n", putanja, err)
			continue
		}

		s := bufio.NewScanner(f)
		for s.Scan() {
			podatak := parseRed(s.Text(), i)
			if podatak == nil {
				continue
			}

			jsonStr, err := json.Marshal(podatak)
			if err != nil {
				log.Printf("[fajl: %v]: json.Marshal() greška: %v\n", putanja, err)
				continue
			}

			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
			if err != nil {
				log.Printf("[fajl: %v]: http.NewRequest() greška: %v\n", putanja, err)
				continue
			}

			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			client.Timeout = time.Second * 15

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("[fajl: %v]: client.Do() greška: %v\n", putanja, err)
				continue
			}

			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("[fajl: %v]: io.ReadAll(resp.Body) greška: %v\n", putanja, err)
			}
			log.Printf("[%v]: %s\n", resp.Status, string(body))

			time.Sleep(time.Second)
		}
		err = s.Err()
		if err != nil {
			log.Print("s.Err() greška: ", err)
		}

		if err = f.Close(); err != nil {
			log.Printf("[fajl: %v]: f.Close() greška: %v\n", putanja, err)
		}
	}

	log.Print("Nema više podataka.")
}

func parseRed(red string, brFajla int) *senzorPodatak {
	podaci := strings.Split(red, ",")
	for i := range podaci {
		podaci[i] = strings.ReplaceAll(podaci[i], " ", "" )
	}

	var (
		podatak senzorPodatak
		temp float64
		vlaznost float64
		pm25 float64
		pm10 float64
		err error
	)

	if brFajla < 12 {
			temp, err = strconv.ParseFloat(podaci[2], 64)
			if err != nil {
				log.Printf("parseRed(), temp: [%v] greška: %v\n", podaci[2], err)
				return nil
			}
			vlaznost, err = strconv.ParseFloat(podaci[3], 64)
			if err != nil {
				log.Printf("parseRed(), vlaznost: [%v] greška: %v\n", podaci[3], err)
				return nil
			}
			pm25, err = strconv.ParseFloat(podaci[4], 64)
			if err != nil {
				log.Printf("parseRed(), pm2.5: [%v] greška: %v\n", podaci[4], err)
				return nil
			}
			pm10, err = strconv.ParseFloat(podaci[5], 64)
			if err != nil {
				log.Printf("parseRed(), pm10: [%v] greška: %v\n", podaci[5], err)
				return nil
			}
	} else {
		temp, err = strconv.ParseFloat(podaci[5], 64)
		if err != nil {
			log.Printf("parseRed(), temp: [%v] greška: %v\n", podaci[5], err)
			return nil
		}
		vlaznost, err = strconv.ParseFloat(podaci[4], 64)
		if err != nil {
			log.Printf("parseRed(), vlaznost: [%v] greška: %v\n", podaci[4], err)
			return nil
		}
		pm25, err = strconv.ParseFloat(podaci[3], 64)
		if err != nil {
			log.Printf("parseRed(), pm2.5: [%v] greška: %v\n", podaci[3], err)
			return nil
		}
		pm10, err = strconv.ParseFloat(podaci[2], 64)
		if err != nil {
			log.Printf("parseRed(), pm10: [%v] greška: %v\n", podaci[2], err)
			return nil
		}
	}

	if math.IsNaN(temp) || math.IsNaN(vlaznost) || math.IsNaN(pm25) || math.IsNaN(pm10) {
		return nil
	}

	podatak.Temperatura = temp
	podatak.VlaznostVazduha = vlaznost
	podatak.Pm25 = pm25
	podatak.Pm10 = pm10

	return &podatak
}
