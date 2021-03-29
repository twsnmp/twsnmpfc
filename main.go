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

	"github.com/labstack/echo/v4"
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

const version = "1000"

func init() {
	flag.StringVar(&dataStorePath, "datastore", "./tmp", "Path to Data Store directory")
	flag.StringVar(&password, "password", "twsnmpfc!", "Master Password")
	flag.StringVar(&port, "port", "8080", "port")
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
	if err = datastore.InitDataStore(ctx, dataStorePath, statikFS); err != nil {
		log.Fatalln(err)
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: "TWSNMP FC起動",
	})
	if err = ping.StartPing(ctx); err != nil {
		log.Fatalln(err)
	}
	if err = report.StartReport(ctx); err != nil {
		log.Fatalln(err)
	}
	if err = logger.StartLogger(ctx); err != nil {
		log.Fatalln(err)
	}
	if err = polling.StartPolling(ctx); err != nil {
		log.Fatalln(err)
	}
	if err = backend.StartBackend(ctx, dataStorePath, version); err != nil {
		log.Fatalln(err)
	}
	if err = notify.StartNotify(ctx); err != nil {
		log.Fatalln(err)
	}
	w := &webapi.WebAPI{
		Statik:        http.FileServer(statikFS),
		Password:      password,
		DataStorePath: dataStorePath,
	}
	e := echo.New()
	webapi.Init(e, w)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		if err := e.Start(":" + port); err != nil {
			log.Println(err)
		}
	}()
	log.Println("Sig")
	sig := <-quit
	log.Println(sig)
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: "TWSNMP FC停止",
	})
	ctxStopWeb, cancelWeb := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelWeb()
	if err := e.Shutdown(ctxStopWeb); err != nil {
		log.Printf("webui shutdown err=%v", err)
	}
	cancel()
	time.Sleep(time.Second * 2)
}
