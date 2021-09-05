package datastore

import (
	"encoding/json"
	"log"

	"go.etcd.io/bbolt"
)

//type=APInfo,ssid=%s,bssid=%s,rssi=%s,Channel=%s,info=%s,count=%d,change=%d,ft=%s,lt=%s
type WifiAPEnt struct {
	ID        string // HOST + BSSID
	Host      string
	BSSID     string
	SSID      string
	RSSI      []RSSIEnt
	Channel   string
	Info      string
	Count     int
	Change    int
	FirstTime int64
	LastTime  int64
}

func GetWifiAP(id string) *WifiAPEnt {
	if v, ok := wifiAP.Load(id); ok {
		return v.(*WifiAPEnt)
	}
	return nil
}

func AddWifiAP(e *WifiAPEnt) {
	wifiAP.Store(e.ID, e)
}

func ForEachWifiAP(f func(*WifiAPEnt) bool) {
	wifiAP.Range(func(k, v interface{}) bool {
		e := v.(*WifiAPEnt)
		return f(e)
	})
}

func loadWifiAP(r *bbolt.Bucket) {
	b := r.Bucket([]byte("wifiAP"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e WifiAPEnt
			if err := json.Unmarshal(v, &e); err == nil {
				wifiAP.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func saveWifiAP(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("wifiAP"))
	wifiAP.Range(func(k, v interface{}) bool {
		e, ok := v.(*WifiAPEnt)
		if !ok {
			return true
		}
		if e.LastTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("Save Report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("Save Report err=%v", err)
		}
		return true
	})
}
