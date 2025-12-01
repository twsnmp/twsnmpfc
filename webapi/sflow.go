package webapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type sFlowFilter struct {
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
	SrcDst    bool
	IP        string
	Port      string
	SrcIP     string
	SrcPort   string
	DstIP     string
	DstPort   string
	Protocol  string
	TCPFlag   string
	Reason    string
	NextTime  int64
	Filter    int
}

type sFlowWebAPI struct {
	Logs     []*sFlowWebAPILogEnt
	NextTime int64
	Process  int
	Filter   int
	Limit    int
}

type sFlowWebAPILogEnt struct {
	Time     int64
	Src      string
	SrcIP    string
	SrcMAC   string
	SrcPort  int
	DstIP    string
	DstMAC   string
	DstPort  int
	Dst      string
	Protocol string
	TCPFlags string
	Bytes    int64
	Reason   int
}

func postSFlow(c echo.Context) error {
	r := new(sFlowWebAPI)
	filter := new(sFlowFilter)
	if err := c.Bind(filter); err != nil {
		return echo.ErrBadRequest
	}
	if filter.SrcDst {
		filter.IP = ""
		filter.Port = ""
	} else {
		filter.SrcIP = ""
		filter.SrcPort = ""
		filter.DstIP = ""
		filter.DstPort = ""
	}
	ipFilter := makeStringFilter(filter.IP)
	srcIPFilter := makeStringFilter(filter.SrcIP)
	dstIPFilter := makeStringFilter(filter.DstIP)
	tcpFlagFilter := makeStringFilter(filter.TCPFlag)
	protocolFilter := makeNumberFilter(filter.Protocol)
	portFilter := makeNumberFilter(filter.Port)
	reasonFilter := makeNumberFilter(filter.Reason)
	srcPortFilter := makeNumberFilter(filter.SrcPort)
	dstPortFilter := makeNumberFilter(filter.DstPort)
	st := makeStartTimeFilter(filter.StartDate, filter.StartTime)
	et := makeEndTimeFilter(filter.EndDate, filter.EndTime)
	if filter.NextTime > 0 {
		et = filter.NextTime
	}
	r.NextTime = 0
	r.Process = 0
	r.Filter = filter.Filter
	i := 0
	to := 15
	if datastore.MapConf.LogTimeout > 0 {
		to = datastore.MapConf.LogTimeout
	}
	end := time.Now().Unix() + int64(to)
	datastore.ForEachLogReverse(st, et, "sflow", func(l *datastore.LogEnt) bool {
		if i > 1000 {
			// 検索期間が15秒を超えた場合
			if time.Now().Unix() > end {
				r.NextTime = l.Time
				return false
			}
			i = 0
		}
		i++
		if len(r.Logs) >= datastore.MapConf.LogDispSize {
			// 検索数が表示件数を超えた場合
			r.NextTime = l.Time
			return false
		}
		r.Process++
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			return true
		}
		var ok bool
		var sp float64
		var sa string
		var dp float64
		var da string
		var prot string
		var tf string
		var bytes float64
		var pi int
		re := new(sFlowWebAPILogEnt)
		if sa, ok = sl["srcAddr"].(string); !ok {
			return true
		}
		if sp, ok = sl["srcPort"].(float64); !ok {
			return true
		}
		if da, ok = sl["dstAddr"].(string); !ok {
			return true
		}
		if dp, ok = sl["dstPort"].(float64); !ok {
			return true
		}
		if prot, ok = sl["protocolStr"].(string); !ok {
			return true
		}
		if v, ok := sl["sourceMacAddress"]; ok {
			if mac, ok := v.(string); ok {
				re.SrcMAC = mac
			}
		}
		if v, ok := sl["destinationMacAddress"]; ok {
			if mac, ok := v.(string); ok {
				re.DstMAC = mac
			}
		}
		if v, ok := sl["discardedReason"]; ok {
			if reason, ok := v.(float64); ok {
				re.Reason = int(reason)
			}
		}
		if v, ok := sl["protocol"]; ok {
			pi = int(v.(float64))
			if prot == "" {
				switch pi {
				case 1:
					prot = "icmp"
				case 2:
					prot = "igmp"
				case 6:
					prot = "tcp"
				case 17:
					prot = "udp"
				default:
					prot = fmt.Sprintf("%d", int(pi))
				}
			}
		}
		if tf, ok = sl["tcpflagsStr"].(string); !ok {
			return true
		}
		if bytes, ok = sl["bytes"].(float64); !ok {
			return true
		}
		spi := int(sp)
		dpi := int(dp)
		if prot == "icmp" {
			spi = dpi / 256
			dpi = dpi % 256
			re.Src = sa
			re.Dst = da
			re.Protocol = fmt.Sprintf("icmp %d/%d", spi, dpi)
		} else {
			re.Src = fmt.Sprintf("%s:%d", sa, spi)
			re.Dst = fmt.Sprintf("%s:%d", da, dpi)
			re.Protocol = prot
		}
		re.Time = l.Time
		re.SrcIP = sa
		re.DstIP = da
		re.SrcPort = spi
		re.DstPort = dpi
		re.TCPFlags = tf
		re.Bytes = int64(bytes)
		if filter.SrcDst {
			if srcIPFilter != nil && !srcIPFilter.Match([]byte(re.SrcIP)) {
				return true
			}
			if dstIPFilter != nil && !dstIPFilter.Match([]byte(re.DstIP)) {
				return true
			}
			if srcPortFilter > 0 && spi != srcPortFilter {
				return true
			}
			if dstPortFilter > 0 && dpi != dstPortFilter {
				return true
			}
		} else {
			if ipFilter != nil && !ipFilter.Match([]byte(re.SrcIP)) && !ipFilter.Match([]byte(re.DstIP)) {
				return true
			}
			if portFilter > 0 && spi != portFilter && dpi != portFilter {
				return true
			}
		}
		if tcpFlagFilter != nil && (pi != 6 || !tcpFlagFilter.Match([]byte(tf))) {
			return true
		}
		if protocolFilter > 0 && pi != protocolFilter {
			return true
		}
		if reasonFilter > 0 && re.Reason != reasonFilter {
			return true
		}
		r.Logs = append(r.Logs, re)
		r.Filter++
		return true
	})
	// 逆順にする
	for i, j := 0, len(r.Logs)-1; i < j; i, j = i+1, j-1 {
		r.Logs[i], r.Logs[j] = r.Logs[j], r.Logs[i]
	}
	r.Limit = datastore.MapConf.LogDispSize
	return c.JSON(http.StatusOK, r)
}

type sFlowCounterFilter struct {
	Remote    string
	Type      string
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
	NextTime  int64
	Filter    int
}

type sFlowCounterWebAPI struct {
	Logs     []*sFlowCounterWebAPILogEnt
	NextTime int64
	Process  int
	Filter   int
	Limit    int
}

type sFlowCounterWebAPILogEnt struct {
	Time int64
	datastore.SFlowCounterEnt
}

func postSFlowCounter(c echo.Context) error {
	r := new(sFlowCounterWebAPI)
	filter := new(sFlowCounterFilter)
	if err := c.Bind(filter); err != nil {
		return echo.ErrBadRequest
	}
	remoteFilter := makeStringFilter(filter.Remote)
	st := makeStartTimeFilter(filter.StartDate, filter.StartTime)
	et := makeEndTimeFilter(filter.EndDate, filter.EndTime)
	if filter.NextTime > 0 {
		et = filter.NextTime
	}
	r.NextTime = 0
	r.Process = 0
	r.Filter = filter.Filter
	i := 0
	to := 15
	if datastore.MapConf.LogTimeout > 0 {
		to = datastore.MapConf.LogTimeout
	}
	end := time.Now().Unix() + int64(to)
	datastore.ForEachLog(st, et, "sflowCounter", func(l *datastore.LogEnt) bool {
		if i > 1000 {
			// 検索期間が15秒を超えた場合
			if time.Now().Unix() > end {
				r.NextTime = l.Time
				return false
			}
			i = 0
		}
		i++
		if len(r.Logs) >= datastore.MapConf.LogDispSize {
			// 検索数が表示件数を超えた場合
			r.NextTime = l.Time
			return false
		}
		r.Process++
		var re sFlowCounterWebAPILogEnt
		if err := json.Unmarshal([]byte(l.Log), &re); err != nil {
			log.Println(err)
			return true
		}
		re.Time = l.Time
		if remoteFilter != nil && !remoteFilter.Match([]byte(re.Remote)) {
			return true
		}
		if filter.Type != "" && filter.Type != re.Type {
			return true
		}
		r.Logs = append(r.Logs, &re)
		r.Filter++
		return true
	})
	// 逆順にする
	for i, j := 0, len(r.Logs)-1; i < j; i, j = i+1, j-1 {
		r.Logs[i], r.Logs[j] = r.Logs[j], r.Logs[i]
	}
	r.Limit = datastore.MapConf.LogDispSize
	return c.JSON(http.StatusOK, r)
}
