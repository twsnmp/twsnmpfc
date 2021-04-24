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
		log.Printf("postNetflow err=%v", err)
		return echo.ErrBadRequest
	}
	srcFilter := makeStringFilter(filter.Src)
	dstFilter := makeStringFilter(filter.Dst)
	protocolFilter := makeStringFilter(filter.Protocol)
	st := makeTimeFilter(filter.StartDate, filter.StartTime, 1)
	et := makeTimeFilter(filter.EndDate, filter.EndTime, 0)
	i := 0
	datastore.ForEachLog(st, et, "netflow", func(l *datastore.LogEnt) bool {
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			log.Printf("postNetflow err=%v", err)
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
		if sa, ok = sl["srcAddr"].(string); !ok {
			log.Printf("postNetflow no srcAddr")
			return true
		}
		if sp, ok = sl["srcPort"].(float64); !ok {
			log.Printf("postNetflow no srcPort")
			return true
		}
		if da, ok = sl["dstAddr"].(string); !ok {
			log.Printf("postNetflow no dstAddr")
			return true
		}
		if dp, ok = sl["dstPort"].(float64); !ok {
			log.Printf("postNetflow no srcPort")
			return true
		}
		if prot, ok = sl["protocolStr"].(string); !ok {
			log.Printf("postNetflow no protocolStr")
			return true
		}
		if prot == "" {
			if v, ok := sl["protocol"]; ok {
				pi := uint8(v.(float64))
				if pi == 1 {
					prot = "icmp"
				} else if pi == 2 {
					prot = "igmp"
				} else if pi == 6 {
					prot = "tcp"
				} else if pi == 17 {
					prot = "udp"
				} else {
					prot = fmt.Sprintf("%d", int(pi))
				}
			}
		}
		if tf, ok = sl["tcpflagsStr"].(string); !ok {
			log.Printf("postNetflow no tcpflagsStr")
			return true
		}
		if packets, ok = sl["packets"].(float64); !ok {
			log.Printf("postNetflow no packets")
			return true
		}
		if bytes, ok = sl["bytes"].(float64); !ok {
			log.Printf("postNetflow no bytes")
			return true
		}
		if lt, ok = sl["last"].(float64); !ok {
			log.Printf("postNetflow no srcPort")
			return true
		}
		if ft, ok = sl["first"].(float64); !ok {
			log.Printf("postNetflow no srcPort")
			return true
		}
		spi := int(sp)
		dpi := int(dp)
		if prot == "icmp" {
			spi = dpi / 256
			dpi = dpi % 256
		}
		re.Time = l.Time
		re.Src = fmt.Sprintf("%s:%d", sa, spi)
		re.Dst = fmt.Sprintf("%s:%d", da, dpi)
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

const tcpFlags = "NCEUAPRSF"

func postIPFIX(c echo.Context) error {
	r := []*netflowWebAPI{}
	filter := new(netflowFilter)
	if err := c.Bind(filter); err != nil {
		log.Printf("postIPFIX err=%v", err)
		return echo.ErrBadRequest
	}
	srcFilter := makeStringFilter(filter.Src)
	dstFilter := makeStringFilter(filter.Dst)
	protocolFilter := makeStringFilter(filter.Protocol)
	st := makeTimeFilter(filter.StartDate, filter.StartTime, 1)
	et := makeTimeFilter(filter.EndDate, filter.EndTime, 0)
	i := 0
	datastore.ForEachLog(st, et, "ipfix", func(l *datastore.LogEnt) bool {
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			log.Printf("postIPFIX err=%v", err)
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
		if ft, ok = sl["flowStartSysUpTime"].(float64); !ok {
			log.Printf("postIPFIX no flowStartSysUpTime")
			return true
		}
		if lt, ok = sl["flowEndSysUpTime"].(float64); !ok {
			log.Printf("postIPFIX no flowEndSysUpTime")
			return true
		}
		if sa, ok = sl["sourceIPv4Address"].(string); !ok {
			if sa, ok = sl["sourceIPv6Address"].(string); !ok {
				log.Printf("postIPFIX no srcAddr")
				return true
			}
		}
		if da, ok = sl["destinationIPv4Address"].(string); !ok {
			if da, ok = sl["destinationIPv6Address"].(string); !ok {
				log.Printf("postIPFIX no dstAddr")
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
				log.Printf("postIPFIX no sourceTransportPort")
				return true
			}
			if dp, ok = sl["destinationTransportPort"].(float64); !ok {
				log.Printf("postIPFIX no destinationTransportPort")
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
		} else {
			log.Println("no data")
		}
		if packets, ok = sl["packetDeltaCount"].(float64); !ok {
			log.Printf("postIPFIX no packetDeltaCount")
			return true
		}
		if bytes, ok = sl["octetDeltaCount"].(float64); !ok {
			log.Printf("postIPFIX no octetDeltaCount")
			return true
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
