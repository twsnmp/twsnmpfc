// Package backend : 裏方の処理
package backend

import (
	"context"
	"log"
)

var (
	versionCheckState int
	versionNum        string
	yasumiMap         map[string]bool
	dspath            string
)

func Start(ctx context.Context, dsp, vn string) error {
	log.Println("start backend")
	dspath = dsp
	versionNum = vn
	yasumiMap = make(map[string]bool)
	makeYasumiMap()
	go monitor(ctx)
	go mapBackend(ctx)
	go aiBackend(ctx)
	return nil
}

func HasNewVersion() bool {
	return versionCheckState == 2
}
