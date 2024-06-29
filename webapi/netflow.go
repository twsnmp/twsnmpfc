package webapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type netflowFilter struct {
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
	NextTime  int64
	Filter    int
}

type netflowWebAPI struct {
	Logs     []*netflowWebAPILogEnt
	NextTime int64
	Process  int
	Filter   int
	Limit    int
}

type netflowWebAPILogEnt struct {
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
	Packets  int64
	Bytes    int64
	Duration float64
}

func postNetFlow(c echo.Context) error {
	r := new(netflowWebAPI)
	filter := new(netflowFilter)
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
	srcPortFilter := makeNumberFilter(filter.SrcPort)
	dstPortFilter := makeNumberFilter(filter.DstPort)
	st := makeTimeFilter(filter.StartDate, filter.StartTime, 1)
	if filter.NextTime > 0 {
		st = filter.NextTime
	}
	et := makeTimeFilter(filter.EndDate, filter.EndTime, 0)
	r.NextTime = 0
	r.Process = 0
	r.Filter = filter.Filter
	i := 0
	to := 15
	if datastore.MapConf.LogTimeout > 0 {
		to = datastore.MapConf.LogTimeout
	}
	end := time.Now().Unix() + int64(to)
	datastore.ForEachLog(st, et, "netflow", func(l *datastore.LogEnt) bool {
		if i > 1000 {
			// 検索期間が15秒を超えた場合
			if time.Now().Unix() > end {
				r.NextTime = l.Time
				return false
			}
			i = 0
		}
		i++
		if r.Filter >= datastore.MapConf.LogDispSize {
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
		var packets float64
		var ft float64
		var lt float64
		var pi int
		re := new(netflowWebAPILogEnt)
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
		if packets, ok = sl["packets"].(float64); !ok {
			return true
		}
		if bytes, ok = sl["bytes"].(float64); !ok {
			return true
		}
		if lt, ok = sl["last"].(float64); !ok {
			return true
		}
		if ft, ok = sl["first"].(float64); !ok {
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
		re.Packets = int64(packets)
		re.Duration = (lt - ft) / 100.0
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
		r.Logs = append(r.Logs, re)
		r.Filter++
		return true
	})
	r.Limit = datastore.MapConf.LogDispSize
	return c.JSON(http.StatusOK, r)
}

const tcpFlags = "NCEUAPRSF"

func postIPFIX(c echo.Context) error {
	r := new(netflowWebAPI)
	filter := new(netflowFilter)
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
	srcPortFilter := makeNumberFilter(filter.SrcPort)
	dstPortFilter := makeNumberFilter(filter.DstPort)
	st := makeTimeFilter(filter.StartDate, filter.StartTime, 1)
	if filter.NextTime > 0 {
		st = filter.NextTime
	}
	et := makeTimeFilter(filter.EndDate, filter.EndTime, 0)
	r.NextTime = 0
	r.Process = 0
	r.Filter = filter.Filter
	i := 0
	to := 15
	if datastore.MapConf.LogTimeout > 0 {
		to = datastore.MapConf.LogTimeout
	}
	end := time.Now().Unix() + int64(to)
	datastore.ForEachLog(st, et, "ipfix", func(l *datastore.LogEnt) bool {
		if i > 1000 {
			// 検索期間が15秒を超えた場合
			if time.Now().Unix() > end {
				r.NextTime = l.Time
				return false
			}
			i = 0
		}
		i++
		if r.Filter >= datastore.MapConf.LogDispSize {
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
		var packets float64
		var ft float64
		var lt float64
		re := new(netflowWebAPILogEnt)
		if ft, ok = sl["flowStartSysUpTime"].(float64); !ok {
			for _, k := range []string{"flowStartMilliseconds", "flowStartSeconds", "flowStartNanoSeconds"} {
				if st, ok := sl[k].(string); ok {
					if t, err := time.Parse(time.RFC3339Nano, st); err == nil {
						ft = float64(t.UnixMilli() / 10)
						break
					}
				}
			}
		}
		if lt, ok = sl["flowEndSysUpTime"].(float64); !ok {
			for _, k := range []string{"flowEndMilliseconds", "flowEndSeconds", "flowEndNanoSeconds"} {
				if st, ok := sl[k].(string); ok {
					if t, err := time.Parse(time.RFC3339Nano, st); err == nil {
						lt = float64(t.UnixMilli() / 10)
						break
					}
				}
			}
		}
		if sa, ok = sl["sourceIPv4Address"].(string); !ok {
			if sa, ok = sl["sourceIPv6Address"].(string); !ok {
				return true
			}
		}
		if da, ok = sl["destinationIPv4Address"].(string); !ok {
			if da, ok = sl["destinationIPv6Address"].(string); !ok {
				return true
			}
		}
		prot = "unknown"
		sp = 0
		dp = 0
		var icmpTypeCode float64
		var pi float64
		if icmpTypeCode, ok = sl["icmpTypeCodeIPv6"].(float64); ok {
			prot = "icmpv6"
			sp = float64(int(icmpTypeCode) / 256)
			dp = float64(int(icmpTypeCode) % 256)
			pi = 1
		} else if icmpTypeCode, ok = sl["icmpTypeCodeIPv4"].(float64); ok {
			prot = "icmpv4"
			sp = float64(int(icmpTypeCode) / 256)
			dp = float64(int(icmpTypeCode) % 256)
			pi = 1
		} else if pi, ok = sl["protocolIdentifier"].(float64); ok {
			if sp, ok = sl["sourceTransportPort"].(float64); !ok {
				return true
			}
			if dp, ok = sl["destinationTransportPort"].(float64); !ok {
				return true
			}
			if int(pi) == 6 {
				if t, ok := sl["tcpflagsStr"]; !ok {
					var tfb float64
					if tfb, ok = sl["tcpControlBits"].(float64); ok {
						f := uint8(tfb)
						flags := []byte{}
						for i := uint8(0); i < 8; i++ {
							if f&0x01 > 0 {
								flags = append(flags, tcpFlags[8-i])
							} else {
								flags = append(flags, '.')
							}
							f >>= 1
						}
						tf = "[" + string(flags) + "]"
					}
				} else {
					tf = t.(string)
				}
				prot = "tcp"
			} else if int(pi) == 17 {
				prot = "udp"
			} else if int(pi) == 1 {
				prot = "icmp"
			} else {
				if v, ok := sl["protocolStr"]; ok {
					prot = v.(string)
				} else {
					prot = fmt.Sprintf("%d", int(pi))
				}
			}
		}
		if packets, ok = sl["packetDeltaCount"].(float64); !ok {
			return true
		}
		if bytes, ok = sl["octetDeltaCount"].(float64); !ok {
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
		spi := int(sp)
		dpi := int(dp)
		re.Time = l.Time
		if int(pi) == 1 {
			re.Src = sa
			re.Dst = da
			re.Protocol = fmt.Sprintf("%s %d/%d", prot, spi, dpi)
		} else {
			re.Src = fmt.Sprintf("%s:%d", sa, spi)
			re.Dst = fmt.Sprintf("%s:%d", da, dpi)
			re.Protocol = prot
		}
		re.SrcIP = sa
		re.DstIP = da
		re.SrcPort = spi
		re.DstPort = dpi
		re.TCPFlags = tf
		re.Bytes = int64(bytes)
		re.Packets = int64(packets)
		re.Duration = (lt - ft) / 100.0
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
		if protocolFilter > 0 && int(pi) != protocolFilter {
			return true
		}
		r.Logs = append(r.Logs, re)
		r.Filter++
		return true
	})
	r.Limit = datastore.MapConf.LogDispSize
	return c.JSON(http.StatusOK, r)
}
