package datastore

import (
	"log"
	"strings"
)

type GrokEnt struct {
	Pat string
	Ok  string
}

var grokMap = make(map[string]*GrokEnt)

func loadGrokMap(s string) {
	for _, l := range strings.Split(s, "\n") {
		l := strings.TrimSpace(l)
		if len(l) < 1 || strings.HasPrefix(l, "#") {
			continue
		}
		l += "\t"
		e := strings.Split(l, "\t")
		if len(e) < 3 {
			continue
		}
		grokMap[e[0]] = &GrokEnt{
			Pat: e[1],
			Ok:  e[2],
		}
	}
	log.Printf("grok=%#v", grokMap)
}

func GetGrokEnt(k string) *GrokEnt {
	if r, ok := grokMap[k]; ok {
		return r
	}
	return nil
}
