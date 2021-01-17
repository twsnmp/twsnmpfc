package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rakyll/statik/fs"
	_ "github.com/twsnmp/twsnmpfc/statik"

	"github.com/twsnmp/twsnmpfc/backend"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/discover"
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

func main() {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatalln(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	ds := datastore.NewDataStore(dataStorePath, statikFS)
	pi := ping.NewPing(ctx)
	di := discover.NewDiscover(ds, pi)
	rp := report.NewReport(ctx, ds)
	po := polling.NewPolling(ctx, ds, rp, pi)
	be := backend.NewBackEnd(ctx, ds, version)
	w := &webapi.WebAPI{
		DataStore: ds,
		Backend:   be,
		Report:    rp,
		Discover:  di,
		Polling:   po,
		Statik:    http.FileServer(statikFS),
		Password:  password,
	}
	e := echo.New()
	webapi.Init(e, w)
	if err := e.Start(":" + port); err != nil {
		log.Println(err)
	}
	cancel()
}
