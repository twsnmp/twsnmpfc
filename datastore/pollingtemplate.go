package datastore

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
)

type PollingTemplateEnt struct {
	ID        string
	Name      string
	Level     string
	Type      string
	Mode      string
	Params    string
	Filter    string
	Extractor string
	Script    string
	Descr     string
	AutoMode  string
}

var pollingTemplateList = make(map[string]*PollingTemplateEnt)

func GetPollingTemplate(id string) *PollingTemplateEnt {
	if pt, ok := pollingTemplateList[id]; ok {
		return pt
	}
	return nil
}

func ForEachPollingTemplate(f func(*PollingTemplateEnt) bool) {
	for _, pt := range pollingTemplateList {
		if !f(pt) {
			break
		}
	}
}

func loadPollingTemplate(js []byte) error {
	var list []PollingTemplateEnt
	if err := json.Unmarshal(js, &list); err != nil {
		return err
	}
	for i := range list {
		if list[i].ID == "" {
			list[i].ID = getID(&list[i])
			pollingTemplateList[list[i].ID] = &list[i]
		}
	}
	return nil
}

func getID(t *PollingTemplateEnt) string {
	s := t.Name + t.Type + t.Mode + t.Level + t.Params + t.Filter + t.Extractor + t.Script
	h := sha256.New()
	if _, err := h.Write([]byte(s)); err != nil {
		log.Printf("get id err=%v", err)
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
