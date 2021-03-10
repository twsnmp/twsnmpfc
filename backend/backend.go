// Package backend : 裏方の処理
package backend

import (
	"context"
)

var (
	versionCheckState int
	versionNum        string
	yasumiMap         map[string]bool
)

func StartBackend(ctx context.Context, vn string) error {
	versionNum = vn
	yasumiMap = make(map[string]bool)
	makeYasumiMap()
	go mapBackend(ctx)
	go aiBackend(ctx)
	return nil
}

func HasNewVersion() bool {
	return versionCheckState == 2
}
