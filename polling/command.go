package polling

// 外部コマンド実行で監視する。

import (
	"fmt"
	"log"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Songmu/timeout"
	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/vjeantet/grok"
)

func doPollingCmd(pe *datastore.PollingEnt) {
	cmds := splitCmd(pe.Polling)
	if len(cmds) < 3 {
		setPollingError("cmd", pe, fmt.Errorf("no cmd"))
		return
	}
	cmd := cmds[0]
	extractor := cmds[1]
	script := cmds[2]
	vm := otto.New()
	lr := make(map[string]string)
	cl := strings.Split(cmd, " ")
	if len(cl) < 1 {
		setPollingError("cmd", pe, fmt.Errorf("no cmd"))
		return
	}
	tio := &timeout.Timeout{
		Cmd:       exec.Command(cl[0], cl[1:]...),
		Duration:  time.Duration(pe.Timeout) * time.Second,
		KillAfter: 5 * time.Second,
	}
	exitStatus, stdout, stderr, err := tio.Run()
	if err != nil {
		setPollingError("cmd", pe, err)
		return
	}
	lr["lastTime"] = time.Now().Format("2006-01-02T15:04")
	lr["stderr"] = stderr
	lr["exitCode"] = fmt.Sprintf("%d", exitStatus.Code)
	if err := vm.Set("exitCode", exitStatus.Code); err != nil {
		log.Printf("doPollingCmd err=%v", err)
	}
	if err := vm.Set("interval", pe.PollInt); err != nil {
		log.Printf("doPollingCmd err=%v", err)
	}
	pe.LastVal = float64(exitStatus.Code)
	if extractor != "" {
		grokEnt := datastore.GetGrokEnt(extractor)
		if grokEnt == nil {
			log.Printf("No grok pattern Polling=%s", pe.Polling)
			setPollingError("cmd", pe, fmt.Errorf("no grok pattern"))
			return
		}
		g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
		if err := g.AddPattern(extractor, grokEnt.Pat); err != nil {
			log.Printf("doPollingCmd err=%v", err)
		}
		cap := fmt.Sprintf("%%{%s}", extractor)
		values, err := g.Parse(cap, string(stdout))
		if err != nil {
			setPollingError("cmd", pe, err)
			return
		}
		for k, v := range values {
			if err := vm.Set(k, v); err != nil {
				log.Printf("doPollingCmd err=%v", err)
			}
			lr[k] = v
		}
	}
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("cmd", pe, err)
		return
	}
	pe.LastVal = 0.0
	for k, v := range lr {
		if strings.Contains(script, k) {
			if fv, err := strconv.ParseFloat(v, 64); err != nil || !math.IsNaN(fv) {
				pe.LastVal = fv
			}
			break
		}
	}
	pe.LastResult = makeLastResult(lr)
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
		return
	}
	setPollingState(pe, pe.Level)
}
