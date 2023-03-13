// Package notify : 通知処理
package notify

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Songmu/timeout"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func checkExecCmd() {
	if datastore.NotifyConf.ExecCmd == "" {
		return
	}
	execLevel := 3
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		ns := getLevelNum(n.State)
		if execLevel > ns {
			execLevel = ns
			if ns == 0 {
				return false
			}
		}
		return true
	})
	if execLevel != lastExecLevel {
		err := ExecNotifyCmd(datastore.NotifyConf.ExecCmd, execLevel)
		r := ""
		if err != nil {
			log.Printf("execNotifyCmd err=%v", err)
			r = fmt.Sprintf("エラー=%v", err)
		}
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: fmt.Sprintf("外部通知コマンド実行 レベル=%d %s", execLevel, r),
		})
		lastExecLevel = execLevel
	}
}

func ExecNotifyCmd(cmd string, level int) error {
	cl := strings.Split(cmd, " ")
	if len(cl) < 1 {
		return fmt.Errorf("notify ExecCmd is empty")
	}
	if filepath.Base(cl[0]) != cl[0] {
		return fmt.Errorf("notify ExecCmd has path")
	}
	strLevel := fmt.Sprintf("%d", level)
	for i, v := range cl {
		if v == "$level" {
			cl[i] = strLevel
		}
	}
	tio := &timeout.Timeout{
		Duration:  60 * time.Second,
		KillAfter: 5 * time.Second,
	}
	if filepath.Ext(cl[0]) == ".sh" {
		cl[0] = filepath.Join(datastore.GetDataStorePath(), "cmd", filepath.Base(cl[0]))
		tio.Cmd = exec.Command("/bin/sh", "-c", strings.Join(cl, " "))
	} else {
		exe := filepath.Join(datastore.GetDataStorePath(), "cmd", filepath.Base(cl[0]))
		if len(cl) == 1 {
			tio.Cmd = exec.Command(exe)
		} else {
			tio.Cmd = exec.Command(exe, cl[1:]...)
		}
	}
	_, _, _, err := tio.Run()
	return err
}
