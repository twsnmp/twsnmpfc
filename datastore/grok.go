package datastore

import (
	"encoding/json"

	"go.etcd.io/bbolt"
)

type GrokEnt struct {
	ID    string
	Name  string
	Descr string
	Pat   string
	Ok    string
}

var defGrockList = []GrokEnt{
	{
		ID:    "EPSLOGIN",
		Name:  "EPSの認証",
		Descr: "EPSで認証した時のユーザーID、クライアントを抽出",
		Pat:   `Login %{NOTSPACE:stat}: \[%{USER:user}\].+cli %{MAC:client}`,
		Ok:    "OK",
	},
	{
		ID:    "FZLOGIN",
		Name:  "FileZenログイン",
		Descr: "FileZenにログインした時のユーザーID、クラアンとを抽出",
		Pat:   `FileZen: %{IP:client} %{USER:user} "Authentication %{NOTSPACE:stat}`,
		Ok:    "succeeded.",
	},
	{
		ID:    "NAOSLOGIN",
		Name:  "NAOSログイン",
		Descr: "NAOSのログイン",
		Pat:   `Login %{NOTSPACE:stat}: \[.+\] %{USER:user}`,
		Ok:    "Success",
	},
	{
		ID:    "DEVICE",
		Name:  "デバイス情報(ip)",
		Descr: "デバイス情報を取得mac=が先のケース",
		Pat:   `mac=%{MAC:mac}.+ip=%{IP:ip}`,
	},
	{
		ID:    "DEVICER",
		Name:  "デバイス情報(mac)",
		Descr: "デバイス情報を取得ip=が先のケース",
		Pat:   `ip=%{IP:ip}.+mac=%{MAC:mac}`,
	},
	{
		ID:    "WELFFLOW",
		Name:  "WELFフロー",
		Descr: "WELF形式のFWのログからフロー情報を取得",
		Pat:   `src=%{IP:src}:%{BASE10NUM:sport}:.+dst=%{IP:dst}:%{BASE10NUM:dport}:.+proto=%{WORD:prot}.+sent=%{BASE10NUM:sent}.+rcvd=%{BASE10NUM:rcvd}.+spkt=%{BASE10NUM:spkt}.+rpkt=%{BASE10NUM:rpkt}`,
	},
	{
		ID:    "OPENWEATHER",
		Name:  "気象情報",
		Descr: "Open Weatherのサイトから気象データを取得",
		Pat:   `"weather":.+"main":\s*"%{WORD:weather}".+"main":.+"temp":\s*%{BASE10NUM:temp}.+"feels_like":\s*%{BASE10NUM:feels_like}.+"temp_min":\s*%{BASE10NUM:temp_min}.+"temp_max":\s*%{BASE10NUM:temp_max}.+"pressure":\s*%{BASE10NUM:pressure}.+"humidity":\s*%{BASE10NUM:humidity}.+"wind":\s*{"speed":\s*%{BASE10NUM:wind}`,
	},
	{
		ID:    "UPTIME",
		Name:  "負荷(uptime)",
		Descr: "uptimeコマンドの出力から負荷を取得",
		Pat:   `load average: %{BASE10NUM:load1m}, %{BASE10NUM:load5m}, %{BASE10NUM:load15m}`,
	},
	{
		ID:    "SSHLOGIN",
		Name:  "SSHのログイン",
		Descr: "SSHでログインした時のユーザーID、クライアントを取得",
		Pat:   `%{NOTSPACE:stat} (password|publickey) for( invalid user | )%{USER:user} from %{IP:client}`,
		Ok:    "Accepted",
	},
}

var grokMap = make(map[string]*GrokEnt)

func loadGrokMap() {
	loadGrokFromDB()
	if len(grokMap) < 1 {
		// 全部消した時はデフォルトに戻す
		for i := range defGrockList {
			UpdateGrokEnt(&defGrockList[i])
		}
	}
}

func loadGrokFromDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("grok"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var g GrokEnt
				if err := json.Unmarshal(v, &g); err == nil {
					grokMap[g.ID] = &g
				}
				return nil
			})
		}
		return nil
	})
}

func GetGrokEnt(id string) *GrokEnt {
	if r, ok := grokMap[id]; ok {
		return r
	}
	return nil
}

// UpdateGrokEnt : Add or Replace GrokEnt
func UpdateGrokEnt(g *GrokEnt) error {
	s, err := json.Marshal(g)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("grok"))
		return b.Put([]byte(g.ID), s)
	})
	if err != nil {
		return err
	}
	grokMap[g.ID] = g
	return nil
}

func DeleteGrokEnt(id string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("grok"))
		return b.Delete([]byte(id))
	})
	delete(grokMap, id)
	if err != nil {
		return err
	}
	return nil
}

func ForEachGrokEnt(f func(*GrokEnt) bool) {
	for _, g := range grokMap {
		if !f(g) {
			break
		}
	}
}
