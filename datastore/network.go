package datastore

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"go.etcd.io/bbolt"
)

type PortEnt struct {
	ID      string
	Name    string
	Polling string
	Index   string
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
	ArpWatch  bool
	Unmanaged bool
	PortWatch bool
	HPorts    int
	X         int
	Y         int
	W         int
	H         int
	SystemID  string
	Error     string
	LLDP      bool
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
	log.Printf("add network dur=%v", time.Since(st))
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
	networks.Store(n.ID, n)
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("networks"))
		return b.Put([]byte(n.ID), s)
	})
	log.Printf("update network dur=%v", time.Since(st))
	return nil
}

func DeleteNetwok(id string) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := networks.Load(id); !ok {
		return ErrInvalidID
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("networks"))
		return b.Delete([]byte(id))
	})
	networks.Delete(id)
	log.Printf("delete network dur=%v", time.Since(st))
	return nil
}

func GetNetwork(id string) *NetworkEnt {
	if db == nil {
		return nil
	}
	if strings.HasPrefix(id, "NET:") {
		a := strings.SplitN(id, ":", 2)
		id = a[1]
	}
	if v, ok := networks.Load(id); ok {
		if n, ok := v.(*NetworkEnt); ok {
			return n
		}
	}
	return nil
}

// ForEachNetworks : Network毎の処理
func ForEachNetworks(f func(*NetworkEnt) bool) {
	networks.Range(func(_, v interface{}) bool {
		if n, ok := v.(*NetworkEnt); ok {
			return f(n)
		}
		return true
	})
}

// FindNetwork : システムIDと管理IPでNetwrorkを検索する
func FindNetwork(id, ip string) *NetworkEnt {
	var ret *NetworkEnt
	networks.Range(func(_, v interface{}) bool {
		if n, ok := v.(*NetworkEnt); ok {
			if n.SystemID == id {
				ret = n
				return false
			}
			if n.IP == ip {
				ret = n
			}
		}
		return true
	})
	return ret
}

// FindNetworkByIP : 管理IPでNetwrorkを検索する
func FindNetworkByIP(ip string) *NetworkEnt {
	var ret *NetworkEnt
	networks.Range(func(_, v interface{}) bool {
		if n, ok := v.(*NetworkEnt); ok {
			if n.IP == ip {
				ret = n
				return false
			}
		}
		return true
	})
	return ret
}

// 保存する前にサイズを補正する
func checkNetwork(n *NetworkEnt) {
	xMax := 5 // 最小幅は5ポート分
	yMax := 0 // 最小の高さは１ポート分
	for _, p := range n.Ports {
		if xMax < p.X {
			xMax = p.X
		}
		if yMax < p.Y {
			yMax = p.Y
		}
	}
	n.W = (xMax+1)*45 + 20
	n.H = (yMax+1)*55 + MapConf.FontSize + 20
	n.Error = ""
	if n.HPorts < 1 {
		n.HPorts = 24
	}
	if n.SystemID == "" {
		n.SystemID = n.IP
	}
}
