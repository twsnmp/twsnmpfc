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
	ok := false
	for i := 0; !ok && i <= pe.Retry; i++ {
		options := ntp.QueryOptions{Timeout: time.Duration(pe.Timeout) * time.Second}
		r, err := ntp.QueryWithOptions(n.IP, options)
		if err != nil {
			log.Printf("doPollingNTP err=%v", err)
			pe.Result["error"] = fmt.Sprintf("%v", err)
			continue
		}
		pe.Result["rtt"] = float64(r.RTT.Nanoseconds())
		pe.Result["stratum"] = float64(r.Stratum)
		pe.Result["refid"] = float64(r.ReferenceID)
		pe.Result["offset"] = float64(r.ClockOffset.Nanoseconds())
		delete(pe.Result, "error")
		ok = true
	}
	if ok {
		setPollingState(pe, "normal")
		return
	}
	setPollingState(pe, pe.Level)
}
