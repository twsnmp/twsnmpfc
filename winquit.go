//go:build linux || darwin

package main

import "os"

func setWindowsQuit(quit chan os.Signal) {
}
