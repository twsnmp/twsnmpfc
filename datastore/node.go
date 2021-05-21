package datastore

import (
	"encoding/json"
	"strings"
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

func loadMapData() error {
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
				nodes.Store(n.ID, &n)
			}
			return nil
		})
		b = tx.Bucket([]byte("lines"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var l LineEnt
				if err := json.Unmarshal(v, &l); err == nil {
					lines.Store(l.ID, &l)
				}
				return nil
			})
		}
		now := time.Now().UnixNano()
		b = tx.Bucket([]byte("pollings"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var p PollingEnt
				if err := json.Unmarshal(v, &p); err == nil {
					if p.Result == nil {
						p.Result = make(map[string]interface{})
					}
					if p.NextTime < now {
						p.NextTime = now
						now += 1000 * 1000 * 500
					}
					pollings.Store(p.ID, &p)
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
		if _, ok := nodes.Load(n.ID); !ok {
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
	nodes.Store(n.ID, n)
	return nil
}

func UpdateNode(n *NodeEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := nodes.Load(n.ID); !ok {
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
	if n, ok := nodes.Load(nodeID); !ok {
		return ErrInvalidID
	} else {
		nn := n.(*NodeEnt)
		AddEventLog(&EventLogEnt{
			Type:     "user",
			Level:    "info",
			NodeName: nn.Name,
			NodeID:   nn.ID,
			Event:    "ノードを削除しました",
		})
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		return b.Delete([]byte(nodeID))
	})
	nodes.Delete(nodeID)
	delList := []string{}
	pollings.Range(func(k, v interface{}) bool {
		if v.(*PollingEnt).NodeID == nodeID {
			delList = append(delList, k.(string))
		}
		return true
	})
	for _, k := range delList {
		DeletePolling(k)
	}
	return nil
}

func GetNode(nodeID string) *NodeEnt {
	if db == nil {
		return nil
	}
	if n, ok := nodes.Load(nodeID); ok {
		return n.(*NodeEnt)
	}
	return nil
}

func FindNodeFromIP(ip string) *NodeEnt {
	var ret *NodeEnt
	nodes.Range(func(_, p interface{}) bool {
		if p.(*NodeEnt).IP == ip {
			ret = p.(*NodeEnt)
			return false
		}
		return true
	})
	return ret
}

func FindNodeFromMAC(mac string) *NodeEnt {
	var ret *NodeEnt
	nodes.Range(func(_, p interface{}) bool {
		if strings.HasPrefix(p.(*NodeEnt).MAC, mac) {
			ret = p.(*NodeEnt)
			return false
		}
		return true
	})
	return ret
}

func ForEachNodes(f func(*NodeEnt) bool) {
	nodes.Range(func(_, p interface{}) bool {
		return f(p.(*NodeEnt))
	})
}

// SetNodeStateChanged :
func SetNodeStateChanged(id string) {
	lastNodeChanged = time.Now()
	stateChangedNodes.Store(id, true)
}

func DeleteNodeStateChanged(id string) {
	stateChangedNodes.Delete(id)
}

func ForEachStateChangedNodes(f func(string) bool) {
	stateChangedNodes.Range(func(id, _ interface{}) bool {
		return f(id.(string))
	})
}
