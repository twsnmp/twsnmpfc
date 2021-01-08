package datastore

import (
	"strings"
)

type GrokEnt struct {
	Pat string
	Ok  string
}

var (
	grokMap = map[string]*GrokEnt{
		"EPSLOGIN":    {Pat: `Login %{GREEDYDATA:stat}: \[%{USER:user}\].+cli %{MAC:client}`, Ok: "OK"},
		"FZLOGIN":     {Pat: `FileZen: %{IP:client} %{USER:user} "Authentication %{GREEDYDATA:stat}`, Ok: "succeeded."},
		"NAOSLOGIN":   {Pat: `Login %{GREEDYDATA:stat}: \[.+\] %{USER:user}`, Ok: "Success"},
		"DEVICE":      {Pat: `mac=%{MAC:mac}.+ip=%{IP:ip}`},
		"DEVICER":     {Pat: `ip=%{IP:ip}.+mac=%{MAC:mac}`},
		"WELFFLOW":    {Pat: `src=%{IP:src}:%{BASE10NUM:sport}:.+dst=%{IP:dst}:%{BASE10NUM:dport}:.+proto=%{WORD:prot}.+sent=%{BASE10NUM:sent}.+rcvd=%{BASE10NUM:rcvd}`},
		"OPENWEATHER": {Pat: `"weather":.+"main":\s*"%{WORD:weather}".+"main":.+"temp":\s*%{BASE10NUM:temp}.+"feels_like":\s*%{BASE10NUM:feels_like}.+"temp_min":\s*%{BASE10NUM:temp_min}.+"temp_max":\s*%{BASE10NUM:temp_max}.+"pressure":\s*%{BASE10NUM:pressure}.+"humidity":\s*%{BASE10NUM:humidity}.+"wind":\s*{"speed":\s*%{BASE10NUM:wind}`},
		"UPTIME":      {Pat: `load average: %{BASE10NUM:load1m}, %{BASE10NUM:load5m}, %{BASE10NUM:load15m}`},
		"SSHLOGIN":    {Pat: `%{GREEDYDATA:stat} (password|publickey) for( invalid user | )%{USER:user} from %{IP:client}`, Ok: "Accepted"},
	}
)

func (ds *DataStore) loadGrokMap(s string) {
	for _, l := range strings.Split(s, "\n") {
		l := strings.TrimSpace(l)
		if len(l) < 1 || strings.HasPrefix(l, "#") {
			continue
		}
		e := splitGrok(l)
		if len(e) < 3 {
			continue
		}
		grokMap[e[0]] = &GrokEnt{
			Pat: e[1],
			Ok:  e[2],
		}
	}
	// TODO:ここにDBから読み込む処理を追加する。
}

func (ds *DataStore) GetGrokEnt(k string) *GrokEnt {
	if r, ok := grokMap[k]; ok {
		return r
	}
	return nil
}

func splitGrok(p string) []string {
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
