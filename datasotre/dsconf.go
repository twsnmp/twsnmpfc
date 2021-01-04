package datastore

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/twsnmp/twsnmpfc/security"
	"go.etcd.io/bbolt"
)

func initConf() {
	MapConf.Community = "public"
	MapConf.PollInt = 60
	MapConf.Retry = 1
	MapConf.Timeout = 1
	MapConf.LogDispSize = 5000
	MapConf.LogDays = 14
	MapConf.ArpWatchLevel = "info"
	MapConf.AILevel = "info"
	MapConf.AIThreshold = 81
	MapConf.Community = "public"
	DiscoverConf.Retry = 1
	DiscoverConf.Timeout = 1
	NotifyConf.InsecureSkipVerify = true
	NotifyConf.Interval = 60
	NotifyConf.Subject = "TWSNMPからの通知"
	NotifyConf.Level = "none"
	InfluxdbConf.DB = "twsnmp"
}

func loadConfFromDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	bSaveConf := false
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		v := b.Get([]byte("mapConf"))
		if v == nil {
			bSaveConf = true
			return nil
		}
		if err := json.Unmarshal(v, &MapConf); err != nil {
			log.Println(fmt.Sprintf("Unmarshal mapConf from DB error=%v", err))
			return err
		}
		v = b.Get([]byte("discoverConf"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &DiscoverConf); err != nil {
			log.Println(fmt.Sprintf("Unmarshal discoverConf from DB error=%v", err))
			return err
		}
		v = b.Get([]byte("notifyConf"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &NotifyConf); err != nil {
			log.Println(fmt.Sprintf("Unmarshal notifyConf from DB error=%v", err))
			return err
		}
		var p DBBackupParamEnt
		v = b.Get([]byte("backup"))
		if v != nil {
			if err := json.Unmarshal(v, &p); err != nil {
				log.Println(fmt.Sprintf("Unmarshal mainWinbackupdowInfo from DB error=%v", err))
			} else {
				if p.BackupFile != "" && p.Daily {
					DBStats.BackupConfigOnly = p.ConfigOnly
					DBStats.BackupFile = p.BackupFile
					DBStats.BackupDaily = p.Daily
					now := time.Now()
					d := 0
					if now.Hour() > 2 {
						d = 1
					}
					nextBackup = time.Date(now.Year(), now.Month(), now.Day()+d, 3, 0, 0, 0, time.Local).UnixNano()
				}
			}
		}
		v = b.Get([]byte("influxdbConf"))
		if v != nil {
			if err := json.Unmarshal(v, &InfluxdbConf); err != nil {
				log.Println(fmt.Sprintf("Unmarshal influxdbConf from DB error=%v", err))
			}
		}
		v = b.Get([]byte("restAPIConf"))
		if v != nil {
			if err := json.Unmarshal(v, &RestAPIConf); err != nil {
				log.Println(fmt.Sprintf("Unmarshal restAPIConf from DB error=%v", err))
			}
		}
		return nil
	})
	if err == nil && MapConf.PrivateKey == "" {
		initSecurityKey()
	}
	if err == nil && bSaveConf {
		if err := SaveMapConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
		if err := SaveNotifyConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
		if err := SaveDiscoverConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
		if err := SaveInfluxdbConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
	}
	return err
}

func initSecurityKey() {
	key, err := security.GenPrivateKey(4096)
	if err != nil {
		log.Printf("initSecurityKey err=%v", err)
		return
	}
	pubkey, err := security.GetSSHPublicKey(key)
	if err != nil {
		log.Printf("initSecurityKey err=%v", err)
		return
	}
	cert, err := security.GenSelfSignCert(key)
	if err != nil {
		log.Printf("initSecurityKey err=%v", err)
		return
	}
	MapConf.PrivateKey = key
	MapConf.PublicKey = pubkey
	MapConf.TLSCert = cert
	log.Printf("initSecurityKey Public Key=%v", pubkey)
	_ = SaveMapConfToDB()
}

func SaveMapConfToDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(MapConf)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("mapConf"), s)
	})
}

func SaveNotifyConfToDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(NotifyConf)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("notifyConf"), s)
	})
}

func SaveInfluxdbConfToDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(InfluxdbConf)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("influxdbConf"), s)
	})
}

func SaveRestAPIConfToDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(RestAPIConf)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("restAPIConf"), s)
	})
}

func SaveBackupParamToDB(p *DBBackupParamEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("backup"), s)
	})
}

func SaveDiscoverConfToDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(DiscoverConf)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("discoverConf"), s)
	})
}

func loadMapDataFromDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		if b == nil {
			return nil
		}
		_ = b.ForEach(func(k, v []byte) error {
			var n NodeEnt
			if err := json.Unmarshal(v, &n); err == nil {
				Nodes.Store(n.ID, &n)
			}
			return nil
		})
		b = tx.Bucket([]byte("lines"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var l LineEnt
				if err := json.Unmarshal(v, &l); err == nil {
					Lines.Store(l.ID, &l)
				}
				return nil
			})
		}
		b = tx.Bucket([]byte("pollings"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var p PollingEnt
				if err := json.Unmarshal(v, &p); err == nil {
					Pollings.Store(p.ID, &p)
				}
				return nil
			})
		}
		return nil
	})
	return err
}

func AddNode(n *NodeEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	for {
		n.ID = makeKey()
		if _, ok := Nodes.Load(n.ID); !ok {
			break
		}
	}
	s, err := json.Marshal(n)
	if err != nil {
		return err
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		return b.Put([]byte(n.ID), s)
	})
	Nodes.Store(n.ID, n)
	return nil
}

func UpdateNode(n *NodeEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := Nodes.Load(n.ID); !ok {
		return ErrInvalidID
	}
	s, err := json.Marshal(n)
	if err != nil {
		return err
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		return b.Put([]byte(n.ID), s)
	})
	return nil
}

func DeleteNode(nodeID string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := Nodes.Load(nodeID); !ok {
		return ErrInvalidID
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		return b.Delete([]byte(nodeID))
	})
	Nodes.Delete(nodeID)
	delList := []string{}
	Pollings.Range(func(k, v interface{}) bool {
		if v.(*PollingEnt).NodeID == nodeID {
			delList = append(delList, k.(string))
		}
		return true
	})
	for _, k := range delList {
		_ = DeletePolling(k)
	}
	return nil
}

func FindNodeFromIP(ip string) *NodeEnt {
	var ret *NodeEnt
	Nodes.Range(func(_, p interface{}) bool {
		if p.(*NodeEnt).IP == ip {
			ret = p.(*NodeEnt)
			return false
		}
		return true
	})
	return ret
}

func AddLine(l *LineEnt) error {
	for {
		l.ID = makeKey()
		if _, ok := Lines.Load(l.ID); !ok {
			break
		}
	}
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(l)
	if err != nil {
		return err
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Put([]byte(l.ID), s)
	})
	Lines.Store(l.ID, l)
	return nil
}

func UpdateLine(l *LineEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := Lines.Load(l.ID); !ok {
		return ErrInvalidID
	}
	s, err := json.Marshal(l)
	if err != nil {
		return err
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Put([]byte(l.ID), s)
	})
	return nil
}

func DeleteLine(lineID string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := Lines.Load(lineID); !ok {
		return ErrInvalidID
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Delete([]byte(lineID))
	})
	Lines.Delete(lineID)
	return nil
}

// AddPolling : ポーリングを追加する
func AddPolling(p *PollingEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	for {
		p.ID = makeKey()
		if _, ok := Pollings.Load(p.ID); !ok {
			break
		}
	}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		return b.Put([]byte(p.ID), s)
	})
	Pollings.Store(p.ID, p)
	return nil
}

func UpdatePolling(p *PollingEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := Pollings.Load(p.ID); !ok {
		return ErrInvalidID
	}
	p.LastTime = time.Now().UnixNano()
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		return b.Put([]byte(p.ID), s)
	})
	Pollings.Store(p.ID, p)
	return nil
}

func DeletePolling(pollingID string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := Pollings.Load(pollingID); !ok {
		return ErrInvalidID
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		return b.Delete([]byte(pollingID))
	})
	Pollings.Delete(pollingID)
	// Delete lines
	Lines.Range(func(_, p interface{}) bool {
		l := p.(*LineEnt)
		if l.PollingID1 == pollingID || l.PollingID2 == pollingID {
			_ = DeleteLine(l.ID)
		}
		return true
	})
	_ = ClearPollingLog(pollingID)
	_ = DeleteAIReesult(pollingID)
	return nil
}

// GetNodePollings : ノードを指定してポーリングリストを取得する
func GetNodePollings(nodeID string) []PollingEnt {
	ret := []PollingEnt{}
	Pollings.Range(func(_, p interface{}) bool {
		if p.(*PollingEnt).NodeID == nodeID {
			ret = append(ret, *p.(*PollingEnt))
		}
		return true
	})
	return ret
}

// GetPollings : ポーリングリストを取得する
func GetPollings() []PollingEnt {
	ret := []PollingEnt{}
	Pollings.Range(func(_, p interface{}) bool {
		ret = append(ret, *p.(*PollingEnt))
		return true
	})
	return ret
}

func GetMIBModuleList() []string {
	ret := []string{}
	if db == nil {
		return ret
	}
	_ = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mibdb"))
		if b == nil {
			return nil
		}
		_ = b.ForEach(func(k, v []byte) error {
			ret = append(ret, string(k))
			return nil
		})
		return nil
	})
	return ret
}

func GetMIBModule(m string) []byte {
	ret := []byte{}
	if db == nil {
		return ret
	}
	_ = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mibdb"))
		if b == nil {
			return nil
		}
		ret = b.Get([]byte(m))
		return nil
	})
	return ret
}

func PutMIBFileToDB(m, path string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mibdb"))
		return b.Put([]byte(m), d)
	})
}

func DelMIBModuleFromDB(m string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mibdb"))
		return b.Delete([]byte(m))
	})
}

func loadPollingTemplateFromDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingTemplates"))
		if b == nil {
			return nil
		}
		_ = b.ForEach(func(k, v []byte) error {
			var pt PollingTemplateEnt
			if err := json.Unmarshal(v, &pt); err == nil {
				PollingTemplates[pt.ID] = &pt
			}
			return nil
		})
		return nil
	})
	return err
}

func getSha1KeyForTemplate(s string) string {
	h := sha1.New()
	if _, err := h.Write([]byte(s)); err != nil {
		log.Printf("getSha1KeyForTemplate err=%v", err)
	}
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func AddPollingTemplate(pt *PollingTemplateEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	pt.ID = getSha1KeyForTemplate(pt.Name + ":" + pt.Type + ":" + pt.NodeType + ":" + pt.Polling)
	if _, ok := PollingTemplates[pt.ID]; ok {
		return fmt.Errorf("duplicate template")
	}
	s, err := json.Marshal(pt)
	if err != nil {
		return err
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingTemplates"))
		return b.Put([]byte(pt.ID), s)
	})
	PollingTemplates[pt.ID] = pt
	return nil
}

func UpdatePollingTemplate(pt *PollingTemplateEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := PollingTemplates[pt.ID]; !ok {
		return ErrInvalidID
	}
	newID := getSha1KeyForTemplate(pt.Name + ":" + pt.Type + ":" + pt.NodeType + ":" + pt.Polling)
	if newID != pt.ID {
		// 更新後に同じ内容のテンプレートがないか確認する
		if _, ok := PollingTemplates[newID]; ok {
			return fmt.Errorf("duplicate template")
		}
	}
	// 削除してから追加する
	_ = DeletePollingTemplate(pt.ID)
	pt.ID = newID
	return AddPollingTemplate(pt)
}

func DeletePollingTemplate(id string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := PollingTemplates[id]; !ok {
		return ErrInvalidID
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingTemplates"))
		return b.Delete([]byte(id))
	})
	delete(PollingTemplates, id)
	return nil
}
