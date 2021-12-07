package datastore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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
	Interval           int
	Level              string
	Report             bool
	CheckUpdate        bool
	NotifyRepair       bool
	NotifyLowScore     bool
	NotifyNewInfo      bool
	URL                string
	HTMLMail           bool
}

func SaveNotifyConf() error {
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

func LoadMailTemplate(t string) string {
	f := fmt.Sprintf("mail_%s.html", t)
	if r, err := os.Open(filepath.Join(dspath, f)); err == nil {
		b, err := ioutil.ReadAll(r)
		if err == nil {
			return string(b)
		}
	}
	return mailTemplate[t]
}
