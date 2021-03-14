package polling

// 外部コマンド実行で監視する。

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/Songmu/timeout"
	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/vjeantet/grok"
)

func doPollingCmd(pe *datastore.PollingEnt) {
	cmd := pe.Params
	extractor := pe.Extractor
	script := pe.Script
	vm := otto.New()
	pe.Result = make(map[string]interface{})
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
	pe.Result["lastTime"] = time.Now().Format("2006-01-02T15:04")
	pe.Result["stderr"] = stderr
	pe.Result["exitCode"] = exitStatus.Code
	if err := vm.Set("exitCode", exitStatus.Code); err != nil {
		log.Printf("doPollingCmd err=%v", err)
	}
	if err := vm.Set("interval", pe.PollInt); err != nil {
		log.Printf("doPollingCmd err=%v", err)
	}
	if extractor != "" {
		grokEnt := datastore.GetGrokEnt(extractor)
		if grokEnt == nil {
			log.Printf("No grok pattern=%s", extractor)
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
			pe.Result[k] = v
		}
	}
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("cmd", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
		return
	}
	setPollingState(pe, pe.Level)
}
