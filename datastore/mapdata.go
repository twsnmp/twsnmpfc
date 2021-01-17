package datastore

import (
	"encoding/json"
	"time"

	"go.etcd.io/bbolt"
)

type NodeEnt struct {
	ID        string
	Name      string
	Descr     string
	Icon      string
	State     string
	X         int
	Y         int
	IP        string
	MAC       string
	SnmpMode  string
	Community string
	User      string
	Password  string
	PublicKey string
	URL       string
	Type      string
	AddrMode  string
}

type LineEnt struct {
	ID         string
	NodeID1    string
	PollingID1 string
	State1     string
	NodeID2    string
	PollingID2 string
	State2     string
}

type PollingEnt struct {
	ID         string
	Name       string
	NodeID     string
	Type       string
	Polling    string
	Level      string
	PollInt    int
	Timeout    int
	Retry      int
	LogMode    int
	NextTime   int64
	LastTime   int64
	LastResult string
	LastVal    float64
	State      string
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
				ds.nodes.Store(n.ID, &n)
			}
			return nil
		})
		b = tx.Bucket([]byte("lines"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var l LineEnt
				if err := json.Unmarshal(v, &l); err == nil {
					ds.lines.Store(l.ID, &l)
				}
				return nil
			})
		}
		b = tx.Bucket([]byte("pollings"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var p PollingEnt
				if err := json.Unmarshal(v, &p); err == nil {
					ds.pollings.Store(p.ID, &p)
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
		if _, ok := ds.nodes.Load(n.ID); !ok {
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
	ds.nodes.Store(n.ID, n)
	return nil
}

func (ds *DataStore) UpdateNode(n *NodeEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if _, ok := ds.nodes.Load(n.ID); !ok {
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
	if _, ok := ds.nodes.Load(nodeID); !ok {
		return ErrInvalidID
	}
	ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		return b.Delete([]byte(nodeID))
	})
	ds.nodes.Delete(nodeID)
	delList := []string{}
	ds.pollings.Range(func(k, v interface{}) bool {
		if v.(*PollingEnt).NodeID == nodeID {
			delList = append(delList, k.(string))
		}
		return true
	})
	for _, k := range delList {
		ds.DeletePolling(k)
	}
	return nil
}

func (ds *DataStore) GetNode(nodeID string) *NodeEnt {
	if ds.db == nil {
		return nil
	}
	if n, ok := ds.nodes.Load(nodeID); ok {
		return n.(*NodeEnt)
	}
	return nil
}

func (ds *DataStore) FindNodeFromIP(ip string) *NodeEnt {
	var ret *NodeEnt
	ds.nodes.Range(func(_, p interface{}) bool {
		if p.(*NodeEnt).IP == ip {
			ret = p.(*NodeEnt)
			return false
		}
		return true
	})
	return ret
}

func (ds *DataStore) ForEachNodes(f func(*NodeEnt) bool) {
	ds.nodes.Range(func(_, p interface{}) bool {
		return f(p.(*NodeEnt))
	})
}

func (ds *DataStore) AddLine(l *LineEnt) error {
	for {
		l.ID = makeKey()
		if _, ok := ds.lines.Load(l.ID); !ok {
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
	ds.lines.Store(l.ID, l)
	return nil
}

func (ds *DataStore) UpdateLine(l *LineEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if _, ok := ds.lines.Load(l.ID); !ok {
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
	if _, ok := ds.lines.Load(lineID); !ok {
		return ErrInvalidID
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Delete([]byte(lineID))
	})
	ds.lines.Delete(lineID)
	return nil
}

// ForEachLines : Line毎の処理
func (ds *DataStore) ForEachLines(f func(*LineEnt) bool) {
	ds.lines.Range(func(_, v interface{}) bool {
		return f(v.(*LineEnt))
	})
}

// AddPolling : ポーリングを追加する
func (ds *DataStore) AddPolling(p *PollingEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	for {
		p.ID = makeKey()
		if _, ok := ds.pollings.Load(p.ID); !ok {
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
	ds.pollings.Store(p.ID, p)
	return nil
}

func (ds *DataStore) UpdatePolling(p *PollingEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if _, ok := ds.pollings.Load(p.ID); !ok {
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
	ds.pollings.Store(p.ID, p)
	return nil
}

func (ds *DataStore) DeletePolling(pollingID string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if _, ok := ds.pollings.Load(pollingID); !ok {
		return ErrInvalidID
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		return b.Delete([]byte(pollingID))
	})
	ds.pollings.Delete(pollingID)
	// Delete lines
	ds.lines.Range(func(_, p interface{}) bool {
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

// GetPolling : ポーリングを取得する
func (ds *DataStore) GetPolling(id string) *PollingEnt {
	p, _ := ds.pollings.Load(id)
	return p.(*PollingEnt)
}

// ForEachPollings : ポーリング毎の処理
func (ds *DataStore) ForEachPollings(f func(*PollingEnt) bool) {
	ds.pollings.Range(func(_, p interface{}) bool {
		return f(p.(*PollingEnt))
	})
}

// SetNodeStateChanged :
func (ds *DataStore) SetNodeStateChanged(id string) {
	ds.lastNodeChanged = time.Now()
	ds.stateChangedNodes.Store(id, true)
}

func (ds *DataStore) DeleteNodeStateChanged(id string) {
	ds.stateChangedNodes.Delete(id)
}

func (ds *DataStore) ForEachStateChangedNodes(f func(string) bool) {
	ds.stateChangedNodes.Range(func(id, _ interface{}) bool {
		return f(id.(string))
	})
}
