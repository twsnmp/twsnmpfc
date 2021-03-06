package polling

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

// API
type restMapStatusEnt struct {
	High      int
	Low       int
	Warn      int
	Normal    int
	Repair    int
	Unknown   int
	DBSize    int64
	DBSizeStr string
	State     string
}

// TWSNMPへのポーリング
func doPollingTWSNMP(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		setPollingError("twsnmp", pe, fmt.Errorf("node not found"))
		return
	}
	ok := false
	var rTime int64
	var body string
	var err error
	for i := 0; !ok && i <= pe.Retry; i++ {
		startTime := time.Now().UnixNano()
		body, err = doTWSNMPGet(n, pe)
		endTime := time.Now().UnixNano()
		if err != nil {
			continue
		}
		rTime = endTime - startTime
		ok = true
	}
	pe.Result["rtt"] = float64(rTime)
	if ok {
		var ms restMapStatusEnt
		if err := json.Unmarshal([]byte(body), &ms); err != nil {
			setPollingError("twsnmp", pe, err)
			return
		}
		pe.Result["state"] = ms.State
		pe.Result["high"] = float64(ms.High)
		pe.Result["low"] = float64(ms.Low)
		pe.Result["warn"] = float64(ms.Warn)
		pe.Result["normal"] = float64(ms.Normal)
		pe.Result["repair"] = float64(ms.Repair)
		pe.Result["dbsize"] = float64(ms.DBSize)
		setPollingState(pe, ms.State)
		return
	}
	setPollingError("twsnmp", pe, err)
}

func doTWSNMPGet(n *datastore.NodeEnt, pe *datastore.PollingEnt) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pe.Timeout)*time.Second)
	defer cancel()
	url := fmt.Sprintf("https://%s:8192/api/mapstatus", n.IP)
	if n.URL != "" {
		url = n.URL + "/api/mapstatus"
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(n.User, n.Password)
	resp, err := insecureClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
