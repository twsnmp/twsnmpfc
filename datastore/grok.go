package datastore

import (
	"encoding/json"
	"log"
	"time"

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
		Pat:   `Login %{NOTSPACE:stat}: \[(host/)*%{USER:user}\].+cli %{MAC:client}`,
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
	{
		ID:    "TWPCAP_STATS",
		Name:  "TWPCAPの統計情報",
		Descr: "TWPCAPで処理したパケット数などの統計情報を抽出",
		Pat:   `type=Stats,total=%{BASE10NUM:total},count=%{BASE10NUM:count},ps=%{BASE10NUM:ps}`,
	},
	{
		ID:    "TWPCAP_IPTOMAC",
		Name:  "TWPCAPのIPとMACアドレス",
		Descr: "TWPCAPで収集したIPとMACアドレスの情報を抽出",
		Pat:   `type=IPToMAC,ip=%{IP:ip},mac=%{MAC:mac},count=%{BASE10NUM:count},change=%{BASE10NUM:chnage},dhcp=%{BASE10NUM:dhcp}`,
	},
	{
		ID:    "TWPCAP_DNS",
		Name:  "TWPCAPのDNS問い合わせ",
		Descr: "TWPCAPで収集したDNSの問い合わせ情報を抽出",
		Pat:   `type=DNS,sv=%{IP:sv},DNSType=%{WORD:dnsType},Name=%{IPORHOST:name},count=%{BASE10NUM:count},change=%{BASE10NUM:chnage},lcl=%{IP:lastIP},lMAC=%{MAC:lastMAC}`,
	},
	{
		ID:    "TWPCAP_DHCP",
		Name:  "TWPCAPのDHCPサーバー情報",
		Descr: "TWPCAPで収集したDHCPサーバー情報を抽出",
		Pat:   `type=DHCP,sv=%{IP:sv},count=%{BASE10NUM:count},offer=%{BASE10NUM:offer},ack=%{BASE10NUM:ack},nak=%{BASE10NUM:nak}`,
	},
	{
		ID:    "TWPCAP_NTP",
		Name:  "TWPCAPのNTPサーバー情報",
		Descr: "TWPCAPで収集したNTPサーバー情報を抽出",
		Pat:   `type=NTP,sv=%{IP:sv},count=%{BASE10NUM:count},change=%{BASE10NUM:change},lcl=%{IP:client},version=%{BASE10NUM:version},stratum=%{BASE10NUM:stratum},refid=%{WORD:refid}`,
	},
	{
		ID:    "TWPCAP_RADIUS",
		Name:  "TWPCAPのRADIUS通信情報",
		Descr: "TWPCAPで収集したRADIUS通信情報を抽出",
		Pat:   `type=RADIUS,cl=%{IP:client},sv=%{IP:server},count=%{BASE10NUM:count},req=%{BASE10NUM:request},accept=%{BASE10NUM:accept},reject=%{BASE10NUM:reject},challenge=%{BASE10NUM:challenge}`,
	},
	{
		ID:    "TWPCAP_TLSFlow",
		Name:  "TWPCAPのTLS通信情報",
		Descr: "TWPCAPで収集したTLS通信情報を抽出",
		Pat:   `type=TLSFlow,cl=%{IP:client},sv=%{IP:server},serv=%{WORD:service},count=%{BASE10NUM:count},handshake=%{BASE10NUM:handshake},alert=%{BASE10NUM:alert},minver=%{DATA:minver},maxver=%{DATA:maxver},cipher=%{DATA:cipher},ft=`,
	},
}

/*
2021/07/18 06:49:40.040 info:local5 twpcap type=Stats,total=265577589,count=621830,ps=10363.83
2021/07/18 06:25:40.040 info:local5 twpcap type=EtherType,0x0806=1356,0x8899=599,0x0800=2317179,0x88cc=1,0x86dd=5086
2021/07/18 07:14:40.040 info:local5 twpcap type=IPToMAC,ip=240d:2:6306:6700:d048:a63a:bbfc:2dab,mac=48:b0:2d:2e:29:19,count=2941,change=0,dhcp=0,ft=2021-07-17T07:34:42+09:00,lt=2021-07-18T07:14:25+09:00
2021/07/18 09:27:40.040 info:local5 twpcap type=DNS,DNSType=Unknown,Name=token.safebrowsing.apple,count=2,change=1,lastIP=240d:2:6306:6700:225:36ff:feab:7753,lastMAC=00:25:36:ab:77:53,ft=2021-07-18T09:27:15+09:00,lt=2021-07-18T09:27:15+09:00
2021/07/18 09:32:40.040 info:local5 twpcap type=DHCP,ip=192.168.1.1,count=246,offer=10,ack=236,nak=0,ft=2021-07-17T07:41:16+09:00,lt=2021-07-18T09:24:04+09:00
2021/07/18 09:38:40.040 info:local5 twpcap type=NTP,ip=17.253.114.125,count=1,change=0,client=192.168.1.9,version=4,stratum=1,refid=0x53484d00,ft=2021-07-18T09:15:51+09:00,lt=2021-07-18T09:15:51+09:00
2021-07-18T00:51:56+00:00  twpcap: type=RADIUS,cl=10.30.1.102,sv=10.30.2.67,count=560,req=272,accept=32,reject=0,challenge=256,ft=2021-07-17T21:00:20Z,lt=2021-07-18T00:39:39Z
2021-07-18T00:49:56+00:00  twpcap: type=TLSFlow,cl=10.30.175.1,sv=35.171.172.30,serv=HTTPS,count=7,handshake=6,alert=0,minver=TLS 1.2,maxver=TLS 1.2,cipher=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,ft=2021-07-18T00:38:20Z,lt=2021-07-18T00:38:20Z
*/

var grokMap = make(map[string]*GrokEnt)

func loadGrokMap() {
	loadGrokFromDB()
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
	st := time.Now()
	err = db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("grok"))
		return b.Put([]byte(g.ID), s)
	})
	if err != nil {
		return err
	}
	grokMap[g.ID] = g
	log.Printf("UpdateGrokEnt dur=%v", time.Since(st))
	return nil
}

func DeleteGrokEnt(id string) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	err := db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("grok"))
		return b.Delete([]byte(id))
	})
	delete(grokMap, id)
	if err != nil {
		return err
	}
	log.Printf("DeleteGrokEnt dur=%v", time.Since(st))
	return nil
}

func ForEachGrokEnt(f func(*GrokEnt) bool) {
	for _, g := range grokMap {
		if !f(g) {
			break
		}
	}
}

func LoadDefGrokEnt() {
	for i := range defGrockList {
		UpdateGrokEnt(&defGrockList[i])
	}
}
