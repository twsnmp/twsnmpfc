// Package backend : 裏方の処理
package backend

import (
	"context"

	"github.com/twsnmp/twsnmpfc/datastore"
)

type Backend struct {
	ds                *datastore.DataStore
	versionCheckState int
	versionNum        string
	yasumiMap         map[string]bool
	aiDone            chan bool
	checkAIMap        map[string]int64
}

func NewBackEnd(ctx context.Context, ds *datastore.DataStore, versionNum string) *Backend {
	b := &Backend{
		ds:         ds,
		versionNum: versionNum,
		yasumiMap:  make(map[string]bool),
		checkAIMap: make(map[string]int64),
		aiDone:     make(chan bool),
	}
	go b.mapBackend(ctx)
	b.makeYasumiMap()
	go b.aiBackend(ctx)
	return b
}

func (b *Backend) HasNewVersion() bool {
	return b.versionCheckState == 2
}
