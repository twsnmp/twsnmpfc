package datastore

import (
	"encoding/json"
	"fmt"
	"log"
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
	IPv6      string
	MAC       string
	SnmpMode  string
	Community string
	User      string
	Password  string
	PublicKey string
	URL       string
	Type      string
	AddrMode  string
	AutoAck   bool
}

type DrawItemType int

const (
	DrawItemTypeRect = iota
	DrawItemTypeEllipse
	DrawItemTypeText
	DrawItemTypeImage
	DrawItemTypePollingText
	DrawItemTypePollingGauge
)

type DrawItemEnt struct {
	ID        string
	Type      DrawItemType
	X         int
	Y         int
	W         int // Width
	H         int // Higeht
	Color     string
	Path      string
	Text      string
	Size      int     // Font Size | GaugeSize
	PollingID string  // Polling ID
	VarName   string  // Pollingから取得する項目
	Format    string  // 表示フォーマット
	Value     float64 // Gaugeの値
	Scale     float64 // 値の補正倍率
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
		b = tx.Bucket([]byte("items"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var di DrawItemEnt
				if err := json.Unmarshal(v, &di); err == nil {
					items.Store(di.ID, &di)
				}
				return nil
			})
		}
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
	setIPv6AndMAC(n)
	s, err := json.Marshal(n)
	if err != nil {
		return err
	}
	st := time.Now()
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		return b.Put([]byte(n.ID), s)
	})
	nodes.Store(n.ID, n)
	log.Printf("AddNode name=%s dur=%v", n.Name, time.Since(st))
	return nil
}

func AddDrawItem(di *DrawItemEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	for {
		di.ID = makeKey()
		if _, ok := items.Load(di.ID); !ok {
			break
		}
	}
	s, err := json.Marshal(di)
	if err != nil {
		return err
	}
	st := time.Now()
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		return b.Put([]byte(di.ID), s)
	})
	items.Store(di.ID, di)
	log.Printf("AddItem  dur=%v", time.Since(st))
	return nil
}

func setIPv6AndMAC(n *NodeEnt) {
	if n.MAC == "" {
		ForEachDevices(func(d *DeviceEnt) bool {
			if d.IP == n.IP {
				mac := d.ID
				v := FindVendor(mac)
				if v != "" {
					mac += fmt.Sprintf("(%s)", v)
				}
				n.MAC = mac
				return false
			}
			return true
		})
	}
	if n.IPv6 == "" && n.MAC != "" {
		ForEachIPReport(func(i *IPReportEnt) bool {
			if strings.HasPrefix(n.MAC, i.MAC) && strings.Contains(i.IP, ":") {
				if n.IPv6 != "" {
					n.IPv6 += ","
					n.IPv6 += i.IP
				}
			}
			return true
		})
	}
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
	DeletePollings(delList)
	return nil
}

func DeleteDrawItem(id string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := items.Load(id); !ok {
		return ErrInvalidID
	} else {
		AddEventLog(&EventLogEnt{
			Type:  "user",
			Level: "info",
			Event: "描画アイテムを削除しました",
		})
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		return b.Delete([]byte(id))
	})
	items.Delete(id)
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

func GetDrawItem(id string) *DrawItemEnt {
	if db == nil {
		return nil
	}
	if di, ok := items.Load(id); ok {
		return di.(*DrawItemEnt)
	}
	return nil
}

func FindNodeFromIP(ip string) *NodeEnt {
	var ret *NodeEnt
	if strings.Contains(ip, ":") {
		// IPv6
		ForEachNodes(func(n *NodeEnt) bool {
			if strings.Contains(n.IPv6, ip) {
				ret = n
				return false
			}
			return true
		})
	} else {
		// IPv4
		ForEachNodes(func(n *NodeEnt) bool {
			if n.IP == ip {
				ret = n
				return false
			}
			return true
		})
	}
	return ret
}

func FindNodeFromMAC(mac string) *NodeEnt {
	var ret *NodeEnt
	if mac == "" {
		return ret
	}
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

func ForEachItems(f func(*DrawItemEnt) bool) {
	items.Range(func(_, p interface{}) bool {
		return f(p.(*DrawItemEnt))
	})
}

func saveAllNodes() error {
	if db == nil {
		return ErrDBNotOpen
	}
	st := time.Now()
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		nodes.Range(func(_, p interface{}) bool {
			pn := p.(*NodeEnt)
			s, err := json.Marshal(pn)
			if err == nil {
				b.Put([]byte(pn.ID), s)
			}
			return true
		})
		b = tx.Bucket([]byte("items"))
		items.Range(func(_, p interface{}) bool {
			di := p.(*DrawItemEnt)
			s, err := json.Marshal(di)
			if err == nil {
				b.Put([]byte(di.ID), s)
			}
			return true
		})
		return nil
	})
	log.Printf("saveAllNodes dur=%v", time.Since(st))
	return nil
}

func CheckNodeAddress(ip, mac, oldmac string) {
	if strings.Contains(ip, ":") {
		// IPv6
		ForEachNodes(func(n *NodeEnt) bool {
			if oldmac != "" && strings.HasPrefix(n.MAC, oldmac) && strings.Contains(n.IPv6, ip) {
				ipv6s := strings.Split(n.IPv6, ",")
				n.IPv6 = ""
				for _, ipv6 := range ipv6s {
					if ipv6 == ip {
						continue
					}
					if n.IPv6 != "" {
						n.IPv6 += ","
					}
					n.IPv6 += ipv6
				}
			}
			if strings.HasPrefix(n.MAC, mac) {
				if !strings.Contains(n.IPv6, ip) {
					if n.IPv6 != "" {
						n.IPv6 += ","
					}
					n.IPv6 += ip
				}
				if oldmac == "" {
					return false
				}
			}
			return true
		})
		return
	}
	// IPv4
	ForEachNodes(func(n *NodeEnt) bool {
		if n.IP == ip {
			if !strings.Contains(n.MAC, mac) {
				v := FindVendor(mac)
				if v != "" {
					mac += fmt.Sprintf("(%s)", v)
				}
				n.MAC = mac
			}
			return false
		}
		return true
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
