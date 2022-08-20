// Package notify : 通知処理
package notify

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

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
	c := filepath.Join(datastore.GetDataStorePath(), "cmd", filepath.Base(cl[0]))
	strLevel := fmt.Sprintf("%d", level)
	if len(cl) == 1 {
		return exec.Command(c).Start()
	}
	for i, v := range cl {
		if v == "$level" {
			cl[i] = strLevel
		}
	}
	return exec.Command(c, cl[1:]...).Start()
}
