package webapi

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/vjeantet/grok"
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
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("抽出設定を更新しました(%s)", g.ID),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteGrok(c echo.Context) error {
	id := c.Param("id")
	if err := datastore.DeleteGrokEnt(id); err != nil {
		log.Printf("deleteGrok err=%v", err)
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("抽出設定を削除しました(%s)", id),
	})
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
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "抽出設定をエクスポートしました",
	})
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
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "抽出設定をインポートしました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

type grockTestEnt struct {
	Pat  string
	Data string
}

type grockTestRespEnt struct {
	ExtractHeader []string
	ExtractDatas  [][]string
}

func postTestGrok(c echo.Context) error {
	gt := new(grockTestEnt)
	if err := c.Bind(gt); err != nil {
		return echo.ErrBadRequest
	}
	grokExtractor, err := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		return echo.ErrBadRequest
	}
	if err = grokExtractor.AddPattern("TEST", gt.Pat); err != nil {
		return echo.ErrBadRequest
	}
	r := new(grockTestRespEnt)
	r.ExtractDatas = [][]string{}
	r.ExtractHeader = []string{}
	for _, l := range strings.Split(gt.Data, "\n") {
		values, err := grokExtractor.Parse("%{TEST}", l)
		if err != nil {
			log.Printf("grock err=%v", err)
			continue
		} else if len(values) > 0 {
			if len(r.ExtractHeader) < 1 {
				for k := range values {
					r.ExtractHeader = append(r.ExtractHeader, k)
					sort.Strings(r.ExtractHeader)
				}
			}
			e := []string{}
			for _, k := range r.ExtractHeader {
				e = append(e, values[k])
			}
			r.ExtractDatas = append(r.ExtractDatas, e)
		}
	}
	return c.JSON(http.StatusOK, r)
}
