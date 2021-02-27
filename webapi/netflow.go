package webapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type netflowFilter struct {
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
	Src       string
	Dst       string
	Protocol  string
	FlowType  string
}

type netflowWebAPI struct {
	Time     int64
	Src      string
	Dst      string
	Protocol string
	TCPFlags string
	Packets  int64
	Bytes    int64
	Duration float64
}

func postNetFlow(c echo.Context) error {
	r := []*netflowWebAPI{}
	filter := new(netflowFilter)
	if err := c.Bind(filter); err != nil {
		log.Printf("postSyslog err=%v", err)
		return echo.ErrBadRequest
	}
	if filter.FlowType != "netflow" && filter.FlowType != "ipfix" {
		filter.FlowType = "netflow"
	}
	srcFilter := makeStringFilter(filter.Src)
	dstFilter := makeStringFilter(filter.Dst)
	protocolFilter := makeStringFilter(filter.Protocol)
	st := makeTimeFilter(filter.StartDate, filter.StartTime, 3)
	et := makeTimeFilter(filter.EndDate, filter.EndTime, 0)
	i := 0
	ipfix := filter.FlowType == "ipfix"
	datastore.ForEachLog(st, et, filter.FlowType, func(l *datastore.LogEnt) bool {
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			log.Printf("postSyslog err=%v", err)
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
		re := new(netflowWebAPI)
		if ipfix {
			if ft, ok = sl["flowStartSysUpTime"].(float64); !ok {
				log.Printf("postSyslog no flowStartSysUpTime")
				return true
			}
			if lt, ok = sl["flowEndSysUpTime"].(float64); !ok {
				log.Printf("postSyslog no flowEndSysUpTime")
				return true
			}
			if sa, ok = sl["sourceIPv4Address"].(string); !ok {
				if sa, ok = sl["sourceIPv6Address"].(string); !ok {
					log.Printf("postSyslog no srcAddr")
					return true
				}
			}
			if da, ok = sl["destinationIPv4Address"].(string); !ok {
				if da, ok = sl["destinationIPv6Address"].(string); !ok {
					log.Printf("postSyslog no dstAddr")
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
			} else if icmpTypeCode, ok = sl["icmpTypeCodeIPv4"].(float64); ok {
				prot = "icmpv4"
				sp = float64(int(icmpTypeCode) / 256)
				dp = float64(int(icmpTypeCode) % 256)
			} else if pi, ok = sl["protocolIdentifier"].(float64); ok {
				if sp, ok = sl["sourceTransportPort"].(float64); !ok {
					log.Printf("postSyslog no sourceTransportPort")
					return true
				}
				if dp, ok = sl["destinationTransportPort"].(float64); !ok {
					log.Printf("postSyslog no destinationTransportPort")
					return true
				}
				if int(pi) == 6 {
					var tfb float64
					if tfb, ok = sl["tcpControlBits"].(float64); ok {
						tf = fmt.Sprintf("%04x", int(tfb))
					}
					prot = "tcp"
				} else if int(pi) == 17 {
					prot = "udp"
				} else {
					prot = fmt.Sprintf("%d", int(pi))
				}
			} else {
				log.Println("no data")
			}
			if packets, ok = sl["packetDeltaCount"].(float64); !ok {
				log.Printf("postSyslog no packetDeltaCount")
				return true
			}
			if bytes, ok = sl["octetDeltaCount"].(float64); !ok {
				log.Printf("postSyslog no octetDeltaCount")
				return true
			}
		} else {
			if sa, ok = sl["srcAddr"].(string); !ok {
				log.Printf("postSyslog no srcAddr")
				return true
			}
			if sp, ok = sl["srcPort"].(float64); !ok {
				log.Printf("postSyslog no srcPort")
				return true
			}
			if da, ok = sl["dstAddr"].(string); !ok {
				log.Printf("postSyslog no dstAddr")
				return true
			}
			if dp, ok = sl["dstPort"].(float64); !ok {
				log.Printf("postSyslog no srcPort")
				return true
			}
			if prot, ok = sl["protocolStr"].(string); !ok {
				log.Printf("postSyslog no protocolStr")
				return true
			}
			if tf, ok = sl["tcpflagsStr"].(string); !ok {
				log.Printf("postSyslog no tcpflagsStr")
				return true
			}
			if packets, ok = sl["packets"].(float64); !ok {
				log.Printf("postSyslog no packets")
				return true
			}
			if bytes, ok = sl["bytes"].(float64); !ok {
				log.Printf("postSyslog no bytes")
				return true
			}
			if lt, ok = sl["last"].(float64); !ok {
				log.Printf("postSyslog no srcPort")
				return true
			}
			if ft, ok = sl["first"].(float64); !ok {
				log.Printf("postSyslog no srcPort")
				return true
			}
		}
		re.Time = l.Time
		re.Src = fmt.Sprintf("%s:%d", sa, int(sp))
		re.Dst = fmt.Sprintf("%s:%d", da, int(dp))
		re.Protocol = prot
		re.TCPFlags = tf
		re.Bytes = int64(bytes)
		re.Packets = int64(packets)
		re.Duration = (lt - ft) / 100.0
		if srcFilter != nil && !srcFilter.Match([]byte(re.Src)) {
			return true
		}
		if dstFilter != nil && !dstFilter.Match([]byte(re.Dst)) {
			return true
		}
		if protocolFilter != nil && !protocolFilter.Match([]byte(re.Protocol)) {
			return true
		}
		r = append(r, re)
		i++
		return i <= datastore.MapConf.LogDispSize
	})
	return c.JSON(http.StatusOK, r)
}
