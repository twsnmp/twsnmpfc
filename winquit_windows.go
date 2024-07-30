//go:build windows

package main

import (
	"os"

	"github.com/containers/winquit/pkg/winquit"
)

func setWindowsQuit(quit chan os.Signal) {
	winquit.SimulateSigTermOnQuit(quit)
}
