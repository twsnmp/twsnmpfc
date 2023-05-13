package datastore

import (
	"log"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

func LoadReport() error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		r := tx.Bucket([]byte("report"))
		loadDevices(r)
		loadUsers(r)
		loadServers(r)
		loadFlows(r)
		loadIPs(r)
		loadEther(r)
		loadDNS(r)
		loadRADIUS(r)
		loadTLS(r)
		loadCert(r)
		loadSensor(r)
		loadWinEventID(r)
		loadWinLogon(r)
		loadWinAccount(r)
		loadWinKerberos(r)
		loadWinPrivilege(r)
		loadWinProcess(r)
		loadWinTask(r)
		loadBlueDevice(r)
		loadEnvMonitor(r)
		loadWifiAP(r)
		loadPowerMonitor(r)
		return nil
	})
}

func SaveReport(last int64) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		saveDevices(b, last)
		saveUsers(b, last)
		saveServers(b, last)
		saveFlows(b, last)
		saveIPs(b, last)
		saveEther(b, last)
		saveDNS(b, last)
		saveRADIUS(b, last)
		saveTLS(b, last)
		saveCert(b, last)
		saveSensor(b, last)
		saveWinEventID(b, last)
		saveWinLogon(b, last)
		saveWinAccount(b, last)
		saveWinKerberos(b, last)
		saveWinPrivilege(b, last)
		saveWinProcess(b, last)
		saveWinTask(b, last)
		saveBlueDevice(b, last)
		saveEnvMonitor(b, last)
		saveWifiAP(b, last)
		savePowerMonitor(b, last)
		log.Printf("SaveReport dur=%v", time.Since(st))
		return nil
	})
}

var reportNameToMap = map[string]*sync.Map{
	"devices":      &devices,
	"users":        &users,
	"servers":      &servers,
	"flows":        &flows,
	"ips":          &ips,
	"ether":        &etherType,
	"dns":          &dnsq,
	"radius":       &radiusFlows,
	"tls":          &tlsFlows,
	"cert":         &certs,
	"sensor":       &sensors,
	"winEventID":   &winEventID,
	"winLogon":     &winLogon,
	"winAccount":   &winAccount,
	"winKerberos":  &winKerberos,
	"winPrivilege": &winPrivilege,
	"winProcess":   &winProcess,
	"winTask":      &winTask,
	"blueDevice":   &blueDevice,
	"envMonitor":   &envMonitor,
	"wifiAP":       &wifiAP,
	"powerMonitor": &powerMonitor,
}

func DeleteReport(report string, ids []string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	st := time.Now()
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			r := b.Bucket([]byte(report))
			if r != nil {
				for _, id := range ids {
					r.Delete([]byte(id))
				}
			}
		}
		return nil
	})
	deleteSyncMap(reportNameToMap[report], ids)
	log.Printf("DeleteReport report=%s len=%d  dur=%v", report, len(ids), time.Since(st))
	return nil
}

func deleteSyncMap(m *sync.Map, ids []string) {
	if m == nil {
		log.Println("delete report err=symc.Map is nil")
		return
	}
	for _, id := range ids {
		m.Delete(id)
	}
}

func ClearReport(r string) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			_ = b.DeleteBucket([]byte(r))
			_, _ = b.CreateBucketIfNotExists([]byte(r))
		}
		return nil
	})
	deleteSyncMapAllData(reportNameToMap[r])
	log.Printf("ClearReport dur=%v", time.Since(st))
	return nil
}

func deleteSyncMapAllData(m *sync.Map) {
	if m == nil {
		log.Println("delete all report err=symc.Map is nil")
		return
	}
	m.Range(func(k, v interface{}) bool {
		m.Delete(k)
		return true
	})
}
