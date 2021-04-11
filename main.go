package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/rakyll/statik/fs"
	"github.com/twsnmp/twsnmpfc/logger"
	"github.com/twsnmp/twsnmpfc/notify"
	_ "github.com/twsnmp/twsnmpfc/statik"

	"github.com/twsnmp/twsnmpfc/backend"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/ping"
	"github.com/twsnmp/twsnmpfc/polling"
	"github.com/twsnmp/twsnmpfc/report"
	"github.com/twsnmp/twsnmpfc/webapi"
)

// From ENV or Command Args
var dataStorePath string
var password string
var port string
var host string
var ip string
var tls bool
var local bool

const version = "1000"

func init() {
	flag.StringVar(&dataStorePath, "datastore", "./datastore", "Path to Data Store directory")
	flag.StringVar(&password, "password", "twsnmpfc!", "Master Password")
	flag.StringVar(&port, "port", "8080", "port")
	flag.StringVar(&host, "host", "", "Host Name for TLS Cert")
	flag.StringVar(&ip, "ip", "", "IP Address for TLS Cert")
	flag.BoolVar(&tls, "tls", false, "Use TLS")
	flag.BoolVar(&local, "local", false, "Local only")
	flag.VisitAll(func(f *flag.Flag) {
		if s := os.Getenv("TWSNMPFC_" + strings.ToUpper(f.Name)); s != "" {
			f.Value.Set(s)
		}
	})
	flag.Parse()
}

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().Format("2006-01-02T15:04:05.999 ") + string(bytes))
}

func main() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
	statikFS, err := fs.New()
	if err != nil {
		log.Fatalln(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err = datastore.Init(ctx, dataStorePath, statikFS); err != nil {
		log.Fatalln(err)
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: "TWSNMP FC起動",
	})
	if err = ping.Start(ctx); err != nil {
		log.Fatalln(err)
	}
	if err = report.Start(ctx); err != nil {
		log.Fatalln(err)
	}
	if err = logger.Start(ctx); err != nil {
		log.Fatalln(err)
	}
	if err = polling.Start(ctx); err != nil {
		log.Fatalln(err)
	}
	if err = backend.Start(ctx, dataStorePath, version); err != nil {
		log.Fatalln(err)
	}
	if err = notify.Start(ctx); err != nil {
		log.Fatalln(err)
	}
	w := &webapi.WebAPI{
		Statik:        http.FileServer(statikFS),
		Port:          port,
		UseTLS:        tls,
		Local:         local,
		IP:            ip,
		Host:          host,
		Password:      password,
		DataStorePath: dataStorePath,
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go webapi.Start(w)
	<-quit
	log.Println("quit by signal")
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: "TWSNMP FC停止",
	})
	webapi.Stop()
	cancel()
	time.Sleep(time.Second * 2)
}
