package datastore

import (
	"encoding/json"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

type PortEnt struct {
	Name    string
	Polling string
	X       int
	Y       int
	State   string
}

type NetworkEnt struct {
	ID        string
	Name      string
	Descr     string
	IP        string // 管理用IP
	SnmpMode  string
	Community string
	User      string
	Password  string
	URL       string
	X         int
	Y         int
	W         int
	H         int
	Error     string
	Ports     []PortEnt
}

func AddNetwork(n *NetworkEnt) error {
	st := time.Now()
	for {
		n.ID = makeKey()
		if _, ok := networks.Load(n.ID); !ok {
			break
		}
	}
	if db == nil {
		return ErrDBNotOpen
	}
	checkNetwork(n)
	s, err := json.Marshal(n)
	if err != nil {
		return err
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("networks"))
		return b.Put([]byte(n.ID), s)
	})
	networks.Store(n.ID, n)
	log.Printf("AddNetwork dur=%v", time.Since(st))
	return nil
}

func UpdateNetwork(n *NetworkEnt) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := networks.Load(n.ID); !ok {
		return ErrInvalidID
	}
	checkNetwork(n)
	s, err := json.Marshal(n)
	if err != nil {
		return err
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("networks"))
		return b.Put([]byte(n.ID), s)
	})
	log.Printf("UpdateNetwork dur=%v", time.Since(st))
	return nil
}

func DeleteNetwok(id string) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	if n, ok := networks.Load(id); !ok {
		return ErrInvalidID
	} else {
		if nn, ok := n.(*NetworkEnt); ok {
			AddEventLog(&EventLogEnt{
				Type:     "user",
				Level:    "info",
				NodeName: nn.Name,
				Event:    "ネットワークを削除しました",
			})
		}
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("networks"))
		return b.Delete([]byte(id))
	})
	networks.Delete(id)
	log.Printf("DeleteNetwork dur=%v", time.Since(st))
	return nil
}

func GetNetwork(id string) *NetworkEnt {
	if db == nil {
		return nil
	}
	if n, ok := networks.Load(id); ok {
		return n.(*NetworkEnt)
	}
	return nil
}

// ForEachNetworks : Network毎の処理
func ForEachNetworks(f func(*NetworkEnt) bool) {
	networks.Range(func(_, v interface{}) bool {
		return f(v.(*NetworkEnt))
	})
}

// 保存する前にサイズを補正する
func checkNetwork(n *NetworkEnt) {
	xMax := 5 // 最小幅は5ポート分
	yMax := 1 // 最小の高さは１ポート分
	for _, p := range n.Ports {
		if xMax < p.X {
			xMax = p.X
		}
		if yMax < p.Y {
			yMax = p.Y
		}
	}
	n.W = xMax*45 + 10
	n.H = yMax*45 + MapConf.FontSize + 15
}
