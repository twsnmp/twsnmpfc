// Package backend : 裏方の処理
package backend

import (
	"context"
	"sync"
)

var (
	versionCheckState int
	versionNum        string
	yasumiMap         map[string]bool
	dspath            string
)

func Start(ctx context.Context, dsp, vn string, wg *sync.WaitGroup) error {
	dspath = dsp
	versionNum = vn
	yasumiMap = make(map[string]bool)
	makeYasumiMap()
	wg.Add(1)
	go monitor(ctx, wg)
	wg.Add(1)
	go mapBackend(ctx, wg)
	wg.Add(1)
	go aiBackend(ctx, wg)
	return nil
}

func HasNewVersion() bool {
	return versionCheckState == 2
}
