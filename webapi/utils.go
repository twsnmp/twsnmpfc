package webapi

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

var logLevelMap = map[string]*regexp.Regexp{
	"high": regexp.MustCompile("high"),
	"low":  regexp.MustCompile("(high|low)"),
	"warn": regexp.MustCompile("(high|low|warn)"),
}

func makeTimeFilter(sd, st string, oh int) int64 {
	if sd == "" {
		return time.Now().Add(-time.Hour * time.Duration(oh)).UnixNano()
	}
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

func splitFilter(p string) []string {
	ret := []string{}
	bInQ := false
	tmp := ""
	for _, c := range p {
		if c == '|' {
			if !bInQ {
				ret = append(ret, strings.TrimSpace(tmp))
				tmp = ""
			}
			continue
		}
		if c == '`' {
			bInQ = !bInQ
		} else {
			tmp += string(c)
		}
	}
	ret = append(ret, strings.TrimSpace(tmp))
	return ret
}

type pipeFilterEnt struct {
	reg *regexp.Regexp
	not bool
}

func makePipeFilter(f string) []pipeFilterEnt {
	var ret []pipeFilterEnt
	if f == "" {
		return ret
	}
	a := splitFilter(f)
	for _, e := range a {
		r, err := regexp.Compile(strings.TrimPrefix(e, "!"))
		if err != nil {
			log.Printf("makeSyslogMsgFilter err=%v", err)
			return ret
		}
		ret = append(ret, pipeFilterEnt{
			reg: r,
			not: strings.HasPrefix(e, "!"),
		})
	}
	return ret
}

func getLogLevelFilter(f string) *regexp.Regexp {
	return logLevelMap[f]
}
