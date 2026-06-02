package datastore

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

type MqttStatEnt struct {
	ID       string `json:"ID"`
	State    string `json:"State"`
	ClientID string `json:"ClientID"`
	Topic    string `json:"Topic"`
	Remote   string `json:"Remote"`
	Count    int    `json:"Count"`
	Bytes    int64  `json:"Bytes"`
	First    int64  `json:"First"`
	Last     int64  `json:"Last"`
}

var mqttStatMap sync.Map

func getMqttStatKey(clientID, topic string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(fmt.Sprintf("%s\t%s", clientID, topic))))
}

func UpdateMqttStat(clientID, remote, topic string, b int) {
	k := getMqttStatKey(clientID, topic)
	if v, ok := mqttStatMap.Load(k); ok {
		if s, ok := v.(*MqttStatEnt); ok {
			s.Bytes += int64(b)
			s.Count++
			s.Remote = remote
			s.Last = time.Now().UnixNano()
		}
		return
	}
	mqttStatMap.Store(k, &MqttStatEnt{
		ID:       k,
		ClientID: clientID,
		Topic:    topic,
		Count:    1,
		Remote:   remote,
		Bytes:    int64(b),
		First:    time.Now().UnixNano(),
		Last:     time.Now().UnixNano(),
	})
}

func ForEachMqttStat(f func(s *MqttStatEnt) bool) {
	warnTime := time.Now().AddDate(0, 0, -1).UnixNano()
	lowTime := time.Now().AddDate(0, 0, -5).UnixNano()
	mqttStatMap.Range(func(key any, value any) bool {
		if s, ok := value.(*MqttStatEnt); ok {
			if s.Last < lowTime {
				s.State = "low"
			} else if s.Last < warnTime {
				s.State = "warn"
			} else {
				s.State = "normal"
			}
			return f(s)
		}
		return true
	})
}

func LoadMqttStat() {
	if db == nil {
		return
	}
	st := time.Now()
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mqttStat"))
		if b == nil {
			return nil
		}
		return b.ForEach(func(k []byte, v []byte) error {
			var s MqttStatEnt
			if err := json.Unmarshal(v, &s); err == nil {
				mqttStatMap.Store(string(k), &s)
			}
			return nil
		})
	})
	log.Printf("load mqtt stat dur=%v", time.Since(st))
}

func SaveMqttStats() {
	if db == nil {
		return
	}
	st := time.Now()
	db.Batch(func(tx *bbolt.Tx) error {
		tx.DeleteBucket([]byte("mqttStat"))
		b, err := tx.CreateBucket([]byte("mqttStat"))
		if b == nil || err != nil {
			return nil
		}
		mqttStatMap.Range(func(key any, value any) bool {
			if k, ok := key.(string); ok {
				if s, ok := value.(*MqttStatEnt); ok {
					if j, err := json.Marshal(s); err == nil {
						b.Put([]byte(k), j)
					}
				}
			}
			return true
		})
		return nil
	})
	log.Printf("save mqtt stat dur=%v", time.Since(st))
}

func DeleteMqttStats(ids []string) {
	for _, id := range ids {
		mqttStatMap.Delete(id)
	}
	SaveMqttStats()
}

func DeleteAllMqttStat() error {
	mqttStatMap.Range(func(key, value any) bool {
		mqttStatMap.Delete(key)
		return true
	})
	if db == nil {
		return nil
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		tx.DeleteBucket([]byte("mqttStat"))
		_, err := tx.CreateBucketIfNotExists([]byte("mqttStat"))
		return err
	})
}

func DeleteOldMqttStats(days int) int {
	if days < 1 {
		return 0
	}
	limit := time.Now().AddDate(0, 0, -days).UnixNano()
	delKeys := []string{}
	mqttStatMap.Range(func(key, value any) bool {
		if s, ok := value.(*MqttStatEnt); ok {
			if s.Last < limit {
				if k, ok := key.(string); ok {
					delKeys = append(delKeys, k)
				}
			}
		}
		return true
	})
	if len(delKeys) > 0 {
		for _, k := range delKeys {
			mqttStatMap.Delete(k)
		}
		SaveMqttStats()
	}
	return len(delKeys)
}
