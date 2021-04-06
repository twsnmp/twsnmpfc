package webapi

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"gopkg.in/yaml.v2"
)

func getGrok(c echo.Context) error {
	r := []*datastore.GrokEnt{}
	datastore.ForEachGrokEnt(func(g *datastore.GrokEnt) bool {
		r = append(r, g)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func postGrok(c echo.Context) error {
	g := new(datastore.GrokEnt)
	if err := c.Bind(g); err != nil {
		return echo.ErrBadRequest
	}
	if err := datastore.UpdateGrokEnt(g); err != nil {
		log.Printf("postGrok err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteGrok(c echo.Context) error {
	id := c.Param("id")
	if err := datastore.DeleteGrokEnt(id); err != nil {
		log.Printf("deleteGrok err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getExportGrok(c echo.Context) error {
	r := []*datastore.GrokEnt{}
	datastore.ForEachGrokEnt(func(g *datastore.GrokEnt) bool {
		r = append(r, g)
		return true
	})
	y, err := yaml.Marshal(r)
	if err != nil {
		log.Printf("getExportGrok err=%v", err)
		return echo.ErrInternalServerError
	}
	return c.Blob(http.StatusOK, "text/yaml", y)
}

func postImportGrok(c echo.Context) error {
	f, err := c.FormFile("file")
	if err != nil {
		log.Printf("postImportGrok err=%v", err)
		return echo.ErrBadRequest
	}
	if f.Size > 1024*1024*20 {
		log.Printf("postImportGrok size over=%v", f)
		return echo.ErrBadRequest
	}
	src, err := f.Open()
	if err != nil {
		log.Printf("postImportGrok err=%v", err)
		return echo.ErrBadRequest
	}
	defer src.Close()
	y, err := ioutil.ReadAll(src)
	if err != nil {
		log.Printf("postImportGrok err=%v", err)
		return echo.ErrBadRequest
	}
	l := []datastore.GrokEnt{}
	err = yaml.Unmarshal(y, &l)
	if err != nil {
		log.Printf("postImportGrok err=%v", err)
		return echo.ErrBadRequest
	}
	for i := range l {
		if err = datastore.UpdateGrokEnt(&l[i]); err != nil {
			log.Printf("postImportGrok err=%v", err)
			return echo.ErrBadRequest
		}
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})

}
