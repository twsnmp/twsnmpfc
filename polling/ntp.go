package polling

import (
	"fmt"
	"log"
	"time"

	"github.com/beevik/ntp"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func doPollingNTP(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		setPollingError("ntp", pe, fmt.Errorf("node not found"))
		return
	}
	lr := make(map[string]string)
	ok := false
	for i := 0; !ok && i <= pe.Retry; i++ {
		options := ntp.QueryOptions{Timeout: time.Duration(pe.Timeout) * time.Second}
		r, err := ntp.QueryWithOptions(n.IP, options)
		if err != nil {
			log.Printf("doPollingNTP err=%v", err)
			lr["error"] = fmt.Sprintf("%v", err)
			continue
		}
		pe.LastVal = float64(r.RTT.Nanoseconds())
		lr["rtt"] = fmt.Sprintf("%f", pe.LastVal)
		lr["stratum"] = fmt.Sprintf("%d", r.Stratum)
		lr["refid"] = fmt.Sprintf("%d", r.ReferenceID)
		lr["offset"] = fmt.Sprintf("%d", r.ClockOffset.Nanoseconds())
		delete(lr, "error")
		ok = true
	}
	pe.LastResult = makeLastResult(lr)
	if ok {
		setPollingState(pe, "normal")
		return
	}
	setPollingState(pe, pe.Level)
}
