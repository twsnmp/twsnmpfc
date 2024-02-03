package datastore

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

type NotifyConfEnt struct {
	MailServer         string
	User               string
	Password           string
	InsecureSkipVerify bool
	MailTo             string
	MailFrom           string
	Subject            string
	AddNodeName        bool
	Interval           int
	Level              string
	Report             bool
	CheckUpdate        bool
	NotifyRepair       bool
	NotifyLowScore     bool
	NotifyNewInfo      bool
	URL                string
	HTMLMail           bool
	ChatType           string
	ChatWebhookURL     string
	ExecCmd            string
	LineToken          string
}

func SaveNotifyConf() error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(NotifyConf)
	if err != nil {
		return err
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("SaveNotifyConf dur=%v", time.Since(st))
		return b.Put([]byte("notifyConf"), s)
	})
}

func LoadMailTemplate(t string) string {
	f := fmt.Sprintf("mail_%s.html", t)
	if r, err := os.Open(filepath.Join(dspath, f)); err == nil {
		b, err := io.ReadAll(r)
		if err == nil {
			return string(b)
		}
	}
	return mailTemplate[t]
}

func SaveNotifySchedule() error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(NotifySchedule)
	if err != nil {
		return err
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("SaveNotifySchedule dur=%v", time.Since(st))
		return b.Put([]byte("notifySchedule"), s)
	})
}
