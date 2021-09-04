package datastore

import (
	"encoding/json"
	"log"

	"go.etcd.io/bbolt"
)

type UserClientEnt struct {
	Total int32
	Ok    int32
}

type UserEnt struct {
	ID           string // User ID + Server
	UserID       string
	Server       string
	ServerName   string
	ServerNodeID string
	ClientMap    map[string]UserClientEnt
	Total        int
	Ok           int
	Score        float64
	ValidScore   bool
	Penalty      int64
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

func GetUser(id string) *UserEnt {
	if v, ok := users.Load(id); ok {
		return v.(*UserEnt)
	}
	return nil
}

func AddUser(u *UserEnt) {
	users.Store(u.ID, u)
}

func ForEachUsers(f func(*UserEnt) bool) {
	users.Range(func(k, v interface{}) bool {
		u := v.(*UserEnt)
		return f(u)
	})
}

func DeleteUser(id string) {
	users.Delete(id)
}

// interna use
func loadUsers(r *bbolt.Bucket) {
	b := r.Bucket([]byte("users"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var u UserEnt
			if err := json.Unmarshal(v, &u); err == nil {
				users.Store(u.ID, &u)
			}
			return nil
		})
	}
}

func saveUsers(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("users"))
	users.Range(func(k, v interface{}) bool {
		u := v.(*UserEnt)
		if u.UpdateTime < last {
			return true
		}
		s, err := json.Marshal(u)
		if err != nil {
			log.Printf("Save Report err=%v", err)
			return true
		}
		err = r.Put([]byte(u.ID), s)
		if err != nil {
			log.Printf("Save Report err=%v", err)
		}
		return true
	})
}
