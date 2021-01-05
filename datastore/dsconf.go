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

func (ds *DataStore) initConf() {
	ds.MapConf.Community = "public"
	ds.MapConf.PollInt = 60
	ds.MapConf.Retry = 1
	ds.MapConf.Timeout = 1
	ds.MapConf.LogDispSize = 5000
	ds.MapConf.LogDays = 14
	ds.MapConf.ArpWatchLevel = "info"
	ds.MapConf.AILevel = "info"
	ds.MapConf.AIThreshold = 81
	ds.MapConf.Community = "public"
	ds.DiscoverConf.Retry = 1
	ds.DiscoverConf.Timeout = 1
	ds.NotifyConf.InsecureSkipVerify = true
	ds.NotifyConf.Interval = 60
	ds.NotifyConf.Subject = "TWSNMPからの通知"
	ds.NotifyConf.Level = "none"
	ds.InfluxdbConf.DB = "twsnmp"
}

func (ds *DataStore) loadConfFromDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	bSaveConf := false
	err := ds.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		v := b.Get([]byte("mapConf"))
		if v == nil {
			bSaveConf = true
			return nil
		}
		if err := json.Unmarshal(v, &ds.MapConf); err != nil {
			log.Println(fmt.Sprintf("Unmarshal mapConf from DB error=%v", err))
			return err
		}
		v = b.Get([]byte("discoverConf"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &ds.DiscoverConf); err != nil {
			log.Println(fmt.Sprintf("Unmarshal discoverConf from DB error=%v", err))
			return err
		}
		v = b.Get([]byte("notifyConf"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &ds.NotifyConf); err != nil {
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
					ds.DBStats.BackupConfigOnly = p.ConfigOnly
					ds.DBStats.BackupFile = p.BackupFile
					ds.DBStats.BackupDaily = p.Daily
					now := time.Now()
					d := 0
					if now.Hour() > 2 {
						d = 1
					}
					ds.nextBackup = time.Date(now.Year(), now.Month(), now.Day()+d, 3, 0, 0, 0, time.Local).UnixNano()
				}
			}
		}
		v = b.Get([]byte("influxdbConf"))
		if v != nil {
			if err := json.Unmarshal(v, &ds.InfluxdbConf); err != nil {
				log.Println(fmt.Sprintf("Unmarshal influxdbConf from DB error=%v", err))
			}
		}
		v = b.Get([]byte("restAPIConf"))
		if v != nil {
			if err := json.Unmarshal(v, &ds.RestAPIConf); err != nil {
				log.Println(fmt.Sprintf("Unmarshal restAPIConf from DB error=%v", err))
			}
		}
		return nil
	})
	if err == nil && ds.MapConf.PrivateKey == "" {
		ds.initSecurityKey()
	}
	if err == nil && bSaveConf {
		if err := ds.SaveMapConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
		if err := ds.SaveNotifyConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
		if err := ds.SaveDiscoverConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
		if err := ds.SaveInfluxdbConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
	}
	return err
}

func (ds *DataStore) initSecurityKey() {
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
	ds.MapConf.PrivateKey = key
	ds.MapConf.PublicKey = pubkey
	ds.MapConf.TLSCert = cert
	log.Printf("initSecurityKey Public Key=%v", pubkey)
	_ = ds.SaveMapConfToDB()
}

func (ds *DataStore) SaveMapConfToDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(ds.MapConf)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("mapConf"), s)
	})
}

func (ds *DataStore) SaveNotifyConfToDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(ds.NotifyConf)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("notifyConf"), s)
	})
}

func (ds *DataStore) SaveInfluxdbConfToDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(ds.InfluxdbConf)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("influxdbConf"), s)
	})
}

func (ds *DataStore) SaveRestAPIConfToDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(ds.RestAPIConf)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("restAPIConf"), s)
	})
}

func (ds *DataStore) SaveBackupParamToDB(p *DBBackupParamEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("backup"), s)
	})
}

func (ds *DataStore) SaveDiscoverConfToDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(ds.DiscoverConf)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("discoverConf"), s)
	})
}

func (ds *DataStore) loadMapDataFromDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	err := ds.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		if b == nil {
			return nil
		}
		_ = b.ForEach(func(k, v []byte) error {
			var n NodeEnt
			if err := json.Unmarshal(v, &n); err == nil {
				ds.Nodes.Store(n.ID, &n)
			}
			return nil
		})
		b = tx.Bucket([]byte("lines"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var l LineEnt
				if err := json.Unmarshal(v, &l); err == nil {
					ds.Lines.Store(l.ID, &l)
				}
				return nil
			})
		}
		b = tx.Bucket([]byte("pollings"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var p PollingEnt
				if err := json.Unmarshal(v, &p); err == nil {
					ds.Pollings.Store(p.ID, &p)
				}
				return nil
			})
		}
		return nil
	})
	return err
}

func (ds *DataStore) AddNode(n *NodeEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	for {
		n.ID = makeKey()
		if _, ok := ds.Nodes.Load(n.ID); !ok {
			break
		}
	}
	s, err := json.Marshal(n)
	if err != nil {
		return err
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		return b.Put([]byte(n.ID), s)
	})
	ds.Nodes.Store(n.ID, n)
	return nil
}

func (ds *DataStore) UpdateNode(n *NodeEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if _, ok := ds.Nodes.Load(n.ID); !ok {
		return ErrInvalidID
	}
	s, err := json.Marshal(n)
	if err != nil {
		return err
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		return b.Put([]byte(n.ID), s)
	})
	return nil
}

func (ds *DataStore) DeleteNode(nodeID string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if _, ok := ds.Nodes.Load(nodeID); !ok {
		return ErrInvalidID
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		return b.Delete([]byte(nodeID))
	})
	ds.Nodes.Delete(nodeID)
	delList := []string{}
	ds.Pollings.Range(func(k, v interface{}) bool {
		if v.(*PollingEnt).NodeID == nodeID {
			delList = append(delList, k.(string))
		}
		return true
	})
	for _, k := range delList {
		_ = ds.DeletePolling(k)
	}
	return nil
}

func (ds *DataStore) FindNodeFromIP(ip string) *NodeEnt {
	var ret *NodeEnt
	ds.Nodes.Range(func(_, p interface{}) bool {
		if p.(*NodeEnt).IP == ip {
			ret = p.(*NodeEnt)
			return false
		}
		return true
	})
	return ret
}

func (ds *DataStore) AddLine(l *LineEnt) error {
	for {
		l.ID = makeKey()
		if _, ok := ds.Lines.Load(l.ID); !ok {
			break
		}
	}
	if ds.db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(l)
	if err != nil {
		return err
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Put([]byte(l.ID), s)
	})
	ds.Lines.Store(l.ID, l)
	return nil
}

func (ds *DataStore) UpdateLine(l *LineEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if _, ok := ds.Lines.Load(l.ID); !ok {
		return ErrInvalidID
	}
	s, err := json.Marshal(l)
	if err != nil {
		return err
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Put([]byte(l.ID), s)
	})
	return nil
}

func (ds *DataStore) DeleteLine(lineID string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if _, ok := ds.Lines.Load(lineID); !ok {
		return ErrInvalidID
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Delete([]byte(lineID))
	})
	ds.Lines.Delete(lineID)
	return nil
}

// AddPolling : ポーリングを追加する
func (ds *DataStore) AddPolling(p *PollingEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	for {
		p.ID = makeKey()
		if _, ok := ds.Pollings.Load(p.ID); !ok {
			break
		}
	}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		return b.Put([]byte(p.ID), s)
	})
	ds.Pollings.Store(p.ID, p)
	return nil
}

func (ds *DataStore) UpdatePolling(p *PollingEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if _, ok := ds.Pollings.Load(p.ID); !ok {
		return ErrInvalidID
	}
	p.LastTime = time.Now().UnixNano()
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		return b.Put([]byte(p.ID), s)
	})
	ds.Pollings.Store(p.ID, p)
	return nil
}

func (ds *DataStore) DeletePolling(pollingID string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if _, ok := ds.Pollings.Load(pollingID); !ok {
		return ErrInvalidID
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		return b.Delete([]byte(pollingID))
	})
	ds.Pollings.Delete(pollingID)
	// Delete lines
	ds.Lines.Range(func(_, p interface{}) bool {
		l := p.(*LineEnt)
		if l.PollingID1 == pollingID || l.PollingID2 == pollingID {
			_ = ds.DeleteLine(l.ID)
		}
		return true
	})
	_ = ds.ClearPollingLog(pollingID)
	_ = ds.DeleteAIReesult(pollingID)
	return nil
}

// GetNodePollings : ノードを指定してポーリングリストを取得する
func (ds *DataStore) GetNodePollings(nodeID string) []PollingEnt {
	ret := []PollingEnt{}
	ds.Pollings.Range(func(_, p interface{}) bool {
		if p.(*PollingEnt).NodeID == nodeID {
			ret = append(ret, *p.(*PollingEnt))
		}
		return true
	})
	return ret
}

// GetPollings : ポーリングリストを取得する
func (ds *DataStore) GetPollings() []PollingEnt {
	ret := []PollingEnt{}
	ds.Pollings.Range(func(_, p interface{}) bool {
		ret = append(ret, *p.(*PollingEnt))
		return true
	})
	return ret
}

func (ds *DataStore) GetMIBModuleList() []string {
	ret := []string{}
	if ds.db == nil {
		return ret
	}
	_ = ds.db.View(func(tx *bbolt.Tx) error {
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

func (ds *DataStore) GetMIBModule(m string) []byte {
	ret := []byte{}
	if ds.db == nil {
		return ret
	}
	_ = ds.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mibdb"))
		if b == nil {
			return nil
		}
		ret = b.Get([]byte(m))
		return nil
	})
	return ret
}

func (ds *DataStore) PutMIBFileToDB(m, path string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mibdb"))
		return b.Put([]byte(m), d)
	})
}

func (ds *DataStore) DelMIBModuleFromDB(m string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mibdb"))
		return b.Delete([]byte(m))
	})
}

func (ds *DataStore) loadPollingTemplateFromDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	err := ds.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingTemplates"))
		if b == nil {
			return nil
		}
		_ = b.ForEach(func(k, v []byte) error {
			var pt PollingTemplateEnt
			if err := json.Unmarshal(v, &pt); err == nil {
				ds.PollingTemplates[pt.ID] = &pt
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

func (ds *DataStore) AddPollingTemplate(pt *PollingTemplateEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	pt.ID = getSha1KeyForTemplate(pt.Name + ":" + pt.Type + ":" + pt.NodeType + ":" + pt.Polling)
	if _, ok := ds.PollingTemplates[pt.ID]; ok {
		return fmt.Errorf("duplicate template")
	}
	s, err := json.Marshal(pt)
	if err != nil {
		return err
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingTemplates"))
		return b.Put([]byte(pt.ID), s)
	})
	ds.PollingTemplates[pt.ID] = pt
	return nil
}

func (ds *DataStore) UpdatePollingTemplate(pt *PollingTemplateEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if _, ok := ds.PollingTemplates[pt.ID]; !ok {
		return ErrInvalidID
	}
	newID := getSha1KeyForTemplate(pt.Name + ":" + pt.Type + ":" + pt.NodeType + ":" + pt.Polling)
	if newID != pt.ID {
		// 更新後に同じ内容のテンプレートがないか確認する
		if _, ok := ds.PollingTemplates[newID]; ok {
			return fmt.Errorf("duplicate template")
		}
	}
	// 削除してから追加する
	_ = ds.DeletePollingTemplate(pt.ID)
	pt.ID = newID
	return ds.AddPollingTemplate(pt)
}

func (ds *DataStore) DeletePollingTemplate(id string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if _, ok := ds.PollingTemplates[id]; !ok {
		return ErrInvalidID
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingTemplates"))
		return b.Delete([]byte(id))
	})
	delete(ds.PollingTemplates, id)
	return nil
}
