// Package backend : 裏方の処理
package backend

import (
	"context"
)

var (
	versionCheckState int
	versionNum        string
	yasumiMap         map[string]bool
	aiDone            chan bool
	checkAIMap        map[string]int64
)

func StartBackend(ctx context.Context, versionNum string) error {
	versionNum = versionNum
	yasumiMap = make(map[string]bool)
	checkAIMap = make(map[string]int64)
	aiDone = make(chan bool)
	makeYasumiMap()
	go mapBackend(ctx)
	go aiBackend(ctx)
	return nil
}

func HasNewVersion() bool {
	return versionCheckState == 2
}
