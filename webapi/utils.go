package webapi

import (
	"fmt"
	"log"
	"regexp"
	"time"
)

var logLevelMap = map[string]*regexp.Regexp{
	"high": regexp.MustCompile("high"),
	"low":  regexp.MustCompile("(high|low)"),
	"warn": regexp.MustCompile("(high|low|warn)"),
}

func makeTimeFilter(sd, st string, oh int) int64 {
	var t time.Time
	var err error
	if t, err = time.Parse("2006-01-02T15:04 MST", fmt.Sprintf("%sT%s JST", sd, st)); err != nil {
		log.Printf("makeTimeFilter err=%v", err)
		t = time.Now().Add(-time.Hour * time.Duration(oh))
	}
	return t.UnixNano()
}

func makeStringFilter(f string) *regexp.Regexp {
	if f == "" {
		return nil
	}
	r, err := regexp.Compile(f)
	if err != nil {
		log.Printf("makeStringFilter err=%v", err)
		return nil
	}
	return r
}

func getLogLevelFilter(f string) *regexp.Regexp {
	return logLevelMap[f]
}
