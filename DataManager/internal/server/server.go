package server;

import (
	"fmt"
	"context"
	"net"
	"log"
	"os"
	"time"
	"encoding/json"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"DataManager/internal/senzorPodaci"
	"DataManager/internal/mqtt"
)

type server struct {
	senzorPodaci.UnimplementedSenzorPodaciServer
}

func (s *server) VratiSenzorPodatak(ctx context.Context, id *wrapperspb.Int32Value) (*senzorPodaci.SenzorPodatak, error) {
	log.Print("Pribavljanje podatka sa id-em: ", id.Value)

	conn, err := pgx.Connect(ctx, pgConnString())
	if (err != nil) {
		log.Print("VratiSenzorPodatak(): Greška prilikom konektovanja sa bazom: ", err)
		return nil, status.Error(status.Code(err), err.Error())
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "select * from senzor_podaci where id=$1", id.Value)
	if err != nil {
		log.Print("VratiSenzorPodatak(): Greška prilikom pribavljanja podatka: ", err)
		return nil, status.Error(status.Code(err), err.Error())
	}

	if !rows.Next() {
		poruka := fmt.Sprintf("Ne postoji podatak sa id: %v", id.Value)
		log.Print(poruka)
		return nil, status.Error(codes.NotFound, poruka)
	}
	var podatak senzorPodaci.SenzorPodatak

	vrednosti, err := rows.Values()
	if err != nil {
		log.Print("VratiSenzorPodatak(): Greška rows.Values(): ", err)
		return nil, status.Error(status.Code(err), err.Error())
	}

	podatak.Id = vrednosti[0].(int32)
	podatak.Vreme = timestamppb.New(vrednosti[1].(time.Time))

	temp, err := vrednosti[2].(pgtype.Numeric).Float64Value()
	if err != nil {
		log.Print("VratiSenzorPodatak(): Greška prilikom konvertovanja temperature: ", err)
		return nil, status.Error(status.Code(err), err.Error())
	}
	podatak.Temperatura = float32(temp.Float64)
	vlaznost, err := vrednosti[3].(pgtype.Numeric).Float64Value()
	if err != nil {
		log.Print("VratiSenzorPodatak(): Greška prilikom konvertovanja vlažnosti vazduha: ", err)
		return nil, status.Error(status.Code(err), err.Error())
	}
	podatak.VlaznostVazduha = float32(vlaznost.Float64)
	pm2_5, err := vrednosti[4].(pgtype.Numeric).Float64Value()
	if err != nil {
		log.Print("VratiSenzorPodatak(): Greška prilikom konvertovanja pm2.5 vrednosti: ", err)
		return nil, status.Error(status.Code(err), err.Error())
	}
	podatak.Pm2_5 = float32(pm2_5.Float64)
	pm10, err := vrednosti[5].(pgtype.Numeric).Float64Value()
	if err != nil {
		log.Print("VratiSenzorPodatak(): Greška prilikom konvertovanja pm10 vrednosti: ", err)
		return nil, status.Error(status.Code(err), err.Error())
	}
	podatak.Pm10 = float32(pm10.Float64)

	log.Printf("Podatak sa id %v uspešno pribavljen.\n", id.Value)
	return &podatak, nil
}

func (s *server) SviSenzorPodaci(_ *emptypb.Empty, stream grpc.ServerStreamingServer[senzorPodaci.SenzorPodatak]) error {
	log.Print("Pribavljanje svih podatka.")

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, pgConnString())
	if (err != nil) {
		log.Print("SviSenzorPodaci(): Greška prilikom konektovanja sa bazom: ", err)
		return status.Error(status.Code(err), err.Error())
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "select * from senzor_podaci WHERE ID > 0")
	if err != nil {
		log.Print("SviSenzorPodaci(): Greška prilikom pribavljanja podatka: ", err)
		return status.Error(status.Code(err), err.Error())
	}

	for rows.Next() {
		var podatak senzorPodaci.SenzorPodatak

		vrednosti, err := rows.Values()
		if err != nil {
			log.Print("SviSenzorPodaci(): Greška rows.Values(): ", err)
			return status.Error(status.Code(err), err.Error())
		}

		podatak.Id = vrednosti[0].(int32)
		podatak.Vreme = timestamppb.New(vrednosti[1].(time.Time))

		temp, err := vrednosti[2].(pgtype.Numeric).Float64Value()
		if err != nil {
			log.Print("SviSenzorPodaci(): Greška prilikom konvertovanja temperature: ", err)
			return status.Error(status.Code(err), err.Error())
		}
		podatak.Temperatura = float32(temp.Float64)
		vlaznost, err := vrednosti[3].(pgtype.Numeric).Float64Value()
		if err != nil {
			log.Print("SviSenzorPodaci(): Greška prilikom konvertovanja vlažnosti vazduha: ", err)
			return status.Error(status.Code(err), err.Error())
		}
		podatak.VlaznostVazduha = float32(vlaznost.Float64)
		pm2_5, err := vrednosti[4].(pgtype.Numeric).Float64Value()
		if err != nil {
			log.Print("SviSenzorPodaci(): Greška prilikom konvertovanja pm2.5 vrednosti: ", err)
			return status.Error(status.Code(err), err.Error())
		}
		podatak.Pm2_5 = float32(pm2_5.Float64)
		pm10, err := vrednosti[5].(pgtype.Numeric).Float64Value()
		if err != nil {
			log.Print("SviSenzorPodaci(): Greška prilikom konvertovanja pm10 vrednosti: ", err)
			return status.Error(status.Code(err), err.Error())
		}
		podatak.Pm10 = float32(pm10.Float64)

		if err := stream.Send(&podatak); err != nil {
			log.Print("SviSenzorPodaci(): Greška prilikom slanja podatka: ", err)
			return status.Error(status.Code(err), err.Error())
		}
	}

	log.Print("Svi podaci uspešno pribavljeni.")
	return nil
}

func (s *server) DodajSenzorPodatak(ctx context.Context, sp *senzorPodaci.SenzorPodatak) (*wrapperspb.BoolValue, error) {
	log.Print("Dodavanje novog podatka.")

	fmt.Printf("sp.Pm2_5: %v\n", sp.Pm2_5)

	conn, err := pgx.Connect(ctx, pgConnString())
	if (err != nil) {
		log.Print("DodajSenzorPodatak(): Greška prilikom konektovanja sa bazom: ", err)
		return wrapperspb.Bool(false), status.Error(status.Code(err), err.Error())
	}
	defer conn.Close(ctx)

	var (
		id int
		vreme time.Time
	)

	query := "INSERT INTO senzor_podaci (temperatura, vlaznost_vazduha, pm2_5, pm10) "
	query += "VALUES ($1, $2, $3, $4) RETURNING id, vreme"

	if err := conn.QueryRow(ctx, query , sp.Temperatura, sp.VlaznostVazduha, sp.Pm2_5, sp.Pm10).Scan(&id, &vreme); err != nil {
		log.Print("DodajSenzorPodatak(): Greška prilikom dodavanja podatka u bazu: ", err)
		return wrapperspb.Bool(false), status.Error(status.Code(err), err.Error())
	}

	log.Printf("Uspešno dodat podatak sa ID: %v\n", id)

	klijent := mqtt.NoviKlijent()
	if klijent != nil {
		tokenSub := klijent.Subscribe("topic/NoviPodaci", 1, nil)
		<- tokenSub.Done()
		if tokenSub.Error() != nil {
			log.Print("DodajSenzorPodatak(): klijent.Subscribe(topic/NoviPodaci) greška: ", tokenSub.Error())
		}

		podatak := struct {
			Id int
			Vreme time.Time
			Temperatura float32
			Vlaznost float32
			Pm2_5 float32
			Pm10 float32
		}{
			Id: id,
			Vreme: vreme,
			Temperatura: sp.Temperatura,
			Vlaznost: sp.VlaznostVazduha,
			Pm2_5: sp.Pm2_5,
			Pm10: sp.Pm10,
		}

		jsonStr, err := json.Marshal(podatak)
		if err != nil {
			log.Print("DodajSenzorPodatak(): json.Marshal(podatak) greška: ", err)
		}

		tokenPub := klijent.Publish("topic/NoviPodaci", 0, false, jsonStr)
		go func() {
			<- tokenPub.Done()
			if tokenPub.Error() != nil {
				log.Print("DodajSenzorPodatak(): klijent.Publish(topic/NoviPodaci) greška: ", tokenPub.Error())
			}
		}()
	}

	return wrapperspb.Bool(true), nil
}

func (s *server) IzmeniSenzorPodatak(ctx context.Context, sp *senzorPodaci.SenzorPodatak) (*wrapperspb.BoolValue, error) {
	log.Print("Izmena podatka sa id: ", sp.Id)

	conn, err := pgx.Connect(ctx, pgConnString())
	if (err != nil) {
		log.Print("IzmeniSenzorPodatak(): Greška prilikom konektovanja sa bazom: ", err)
		return wrapperspb.Bool(false), status.Error(status.Code(err), err.Error())
	}
	defer conn.Close(ctx)

	tag, err := conn.Exec(ctx, "UPDATE senzor_podaci SET temperatura = $1, vlaznost_vazduha = $2, pm2_5 = $3, pm10 = $4 WHERE id = $5", sp.Temperatura, sp.VlaznostVazduha, sp.Pm2_5, sp.Pm10, sp.Id)
	if err != nil {
		log.Print("IzmeniSenzorPodatak(): Greška prilikom izmene podatka: ", err)
		return wrapperspb.Bool(false), status.Error(status.Code(err), err.Error())
	}
	if tag.RowsAffected() != 1 {
		poruka := fmt.Sprintf("Ne postoji podatak sa id: %v", sp.Id)
		log.Print("IzmeniSenzorPodatak(): Greška prilikom brisanja podatka: ", poruka)
		return wrapperspb.Bool(false), status.Error(codes.NotFound, poruka)
	}

	log.Print("Podatak uspešno izmenjen.")
	return wrapperspb.Bool(true), nil
}

func (s *server) IzbrisiSenzorPodatak(ctx context.Context, id *wrapperspb.Int32Value) (*wrapperspb.BoolValue, error) {
	log.Print("Brisanje podatka sa id: ", id.Value)

	conn, err := pgx.Connect(ctx, pgConnString())
	if (err != nil) {
		log.Print("IzbrisiSenzorPodatak(): Greška prilikom konektovanja sa bazom: ", err)
		return wrapperspb.Bool(false), status.Error(status.Code(err), err.Error())
	}
	defer conn.Close(ctx)

	tag, err := conn.Exec(ctx, "DELETE FROM senzor_podaci WHERE id = $1", id.Value)
	if err != nil {
		log.Print("IzbrisiSenzorPodatak(): Greška prilikom brisanja podatka: ", err)
		return wrapperspb.Bool(false), status.Error(status.Code(err), err.Error())
	}
	if tag.RowsAffected() != 1 {
		poruka := fmt.Sprintf("Ne postoji podatak sa id: %v", id.Value)
		log.Print("IzbrisiSenzorPodatak(): Greška prilikom brisanja podatka: ", poruka)
		return wrapperspb.Bool(false), status.Error(codes.NotFound, poruka)
	}

	log.Print("Podatak uspešno obrisan.")
	return wrapperspb.Bool(true), nil
}

func (s *server) SviSenzorPodaciPeriod(vp *senzorPodaci.VremenskiPeriod, stream grpc.ServerStreamingServer[senzorPodaci.SenzorPodatak]) error {
	log.Printf("Pribavljanje svih podatka u periodu izmedju %v i %v.\n", vp.Pocetak.AsTime(), vp.Kraj.AsTime())

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, pgConnString())
	if (err != nil) {
		log.Print("SviSenzorPodaciPeriod(): Greška prilikom konektovanja sa bazom: ", err)
		return status.Error(status.Code(err), err.Error())
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "select * from senzor_podaci WHERE vreme BETWEEN $1 AND $2 AND ID > 0", vp.Pocetak.AsTime(), vp.Kraj.AsTime())
	if err != nil {
		log.Print("SviSenzorPodaciPeriod(): Greška prilikom pribavljanja podatka: ", err)
		return status.Error(status.Code(err), err.Error())
	}

	brojRedova := 0
	for rows.Next() {
		brojRedova++
		var podatak senzorPodaci.SenzorPodatak

		vrednosti, err := rows.Values()
		if err != nil {
			log.Print("SviSenzorPodaciPeriod(): Greška rows.Values(): ", err)
			return status.Error(status.Code(err), err.Error())
		}

		podatak.Id = vrednosti[0].(int32)
		podatak.Vreme = timestamppb.New(vrednosti[1].(time.Time))

		temp, err := vrednosti[2].(pgtype.Numeric).Float64Value()
		if err != nil {
			log.Print("SviSenzorPodaciPeriod(): Greška prilikom konvertovanja temperature: ", err)
			return status.Error(status.Code(err), err.Error())
		}
		podatak.Temperatura = float32(temp.Float64)
		vlaznost, err := vrednosti[3].(pgtype.Numeric).Float64Value()
		if err != nil {
			log.Print("SviSenzorPodaciPeriod(): Greška prilikom konvertovanja vlažnosti vazduha: ", err)
			return status.Error(status.Code(err), err.Error())
		}
		podatak.VlaznostVazduha = float32(vlaznost.Float64)
		pm2_5, err := vrednosti[4].(pgtype.Numeric).Float64Value()
		if err != nil {
			log.Print("SviSenzorPodaciPeriod(): Greška prilikom konvertovanja pm2.5 vrednosti: ", err)
			return status.Error(status.Code(err), err.Error())
		}
		podatak.Pm2_5 = float32(pm2_5.Float64)
		pm10, err := vrednosti[5].(pgtype.Numeric).Float64Value()
		if err != nil {
			log.Print("SviSenzorPodaciPeriod(): Greška prilikom konvertovanja pm10 vrednosti: ", err)
			return status.Error(status.Code(err), err.Error())
		}
		podatak.Pm10 = float32(pm10.Float64)

		if err := stream.Send(&podatak); err != nil {
			log.Print("SviSenzorPodaciPeriod(): Greška prilikom slanja podatka: ", err)
			return status.Error(status.Code(err), err.Error())
		}
	}

	if brojRedova == 0 {
		return status.Error(codes.NotFound, "Nema podata u datom vremenskom periodu.")
	}

	log.Print("Svi podaci uspešno pribavljeni.")
	return nil
}

func Start() {
	if err := godotenv.Load(".env"); err != nil {
		log.Print("Greška prilikom učitavanja .env fajla: ", err)
		return
	}

	lis, err := net.Listen("tcp", "0.0.0.0:" + os.Getenv("DM_PORT"))
	if err != nil {
		log.Fatal("Greska net.Listen: ", err)
	}

	grpcServer := grpc.NewServer()
	senzorPodaci.RegisterSenzorPodaciServer(grpcServer, &server {})

	log.Print("Server osluškuje na portu: ", os.Getenv("DM_PORT"))
	grpcServer.Serve(lis)
}

func pgConnString() string {
	return fmt.Sprintf("postgres://%s:%s@db:%s/%s", os.Getenv("PG_USER"), os.Getenv("PG_PASSWORD"), os.Getenv("PG_PORT"), os.Getenv("PG_DBNAME"))
}
