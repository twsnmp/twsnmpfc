package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rakyll/statik/fs"
	_ "github.com/twsnmp/twsnmpfc/statik"
)

// From ENV or Command Args
var dbPath string
var logLevel string
var masterPassword string

func init() {
	flag.StringVar(&dbPath, "dbpath", "./twsnmpfc.twdb", "Path to DB file")
	flag.StringVar(&logLevel, "loglevel", "info", "Log Level")
	flag.StringVar(&masterPassword, "password", "twsnmpfc!", "Master Password")
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
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	h := http.FileServer(statikFS)
	e.GET("/*", echo.WrapHandler(http.StripPrefix("/", h)))

	e.Logger.Fatal(e.Start(":8080"))
}
