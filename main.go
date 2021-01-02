package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rakyll/statik/fs"
	_ "github.com/twsnmp/twsnmpfc/statik"
)

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
