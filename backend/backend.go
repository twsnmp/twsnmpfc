// Package backend : 裏方の処理
package backend

import (
	"context"

	"github.com/twsnmp/twsnmpfc/datastore"
)

type BackEnd struct {
	ds                *datastore.DataStore
	versionCheckState int
	versionNum        string
}

func NewBackEnd(ctx context.Context, ds *datastore.DataStore, versionNum string) *BackEnd {
	b := &BackEnd{
		ds:         ds,
		versionNum: versionNum,
	}
	go b.mapBackend(ctx)
	return b
}

func (b *BackEnd) HasNewVersion() bool {
	return b.versionCheckState == 2
}
