package logger

/*
  syslog,tarp,netflow5,ipfixをログに記録する
*/

import (
	"bytes"
	"encoding/json"
	"log"

	"fmt"
	"net"
	"strings"
	"time"

	"github.com/tehmaze/netflow"
	"github.com/tehmaze/netflow/ipfix"
	"github.com/tehmaze/netflow/netflow5"
	"github.com/tehmaze/netflow/read"
	"github.com/tehmaze/netflow/session"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
)

func netflowd(stopCh chan bool) {
	var readSize = 2 << 16
	var addr *net.UDPAddr
	var err error
	log.Printf("netflowd start")
	if addr, err = net.ResolveUDPAddr("udp", ":2055"); err != nil {
		log.Printf("netflowd err=%v", err)
		return
	}
	var server *net.UDPConn
	if server, err = net.ListenUDP("udp", addr); err != nil {
		log.Printf("netflowd err=%v", err)
		return
	}
	defer server.Close()
	if err = server.SetReadBuffer(readSize); err != nil {
		log.Printf("netflowd err=%v", err)
		return
	}
	decoders := make(map[string]*netflow.Decoder)
	buf := make([]byte, 8192)
	for {
		select {
		case <-stopCh:
			{
				log.Printf("netflowd stop")
				return
			}
		default:
			{
				_ = server.SetReadDeadline(time.Now().Add(time.Second * 2))
				var remote *net.UDPAddr
				var octets int
				if octets, remote, err = server.ReadFromUDP(buf); err != nil {
					if !strings.Contains(err.Error(), "timeout") {
						log.Printf("netflowd err=%v", err)
					}
					continue
				}
				d, found := decoders[remote.String()]
				if !found {
					s := session.New()
					d = netflow.NewDecoder(s)
					decoders[remote.String()] = d
				}
				m, err := d.Read(bytes.NewBuffer(buf[:octets]))
				if err != nil {
					log.Printf("netflowd err=%v", err)
					continue
				}
				switch p := m.(type) {
				case *netflow5.Packet:
					{
						logNetflow(p)
						report.UpdateFlowSensor(remote.IP.String(), "netflow", len(p.Records))
					}
				case *ipfix.Message:
					{
						r := logIPFIX(p)
						report.UpdateFlowSensor(remote.IP.String(), "ipfix", r)
					}
				}
			}
		}
	}
}

func logIPFIX(p *ipfix.Message) int {
	r := 0
	for _, ds := range p.DataSets {
		if ds.Records == nil {
			continue
		}
		for _, dr := range ds.Records {
			r++
			var record = make(map[string]interface{})
			for _, f := range dr.Fields {
				if f.Translated != nil {
					if f.Translated.Name != "" {
						record[f.Translated.Name] = f.Translated.Value
						if f.Translated.Name == "protocolIdentifier" {
							record["protocolStr"] = read.Protocol(f.Translated.Value.(uint8))
						}
						if f.Translated.Name == "tcpControlBits" {
							record["tcpflagsStr"] = read.TCPFlags(uint8(f.Translated.Value.(uint16)))
						}
					} else {
						record[fmt.Sprintf("%d.%d", f.Translated.EnterpriseNumber, f.Translated.InformationElementID)] = f.Bytes
					}
				} else {
					record["raw"] = f.Bytes
				}
			}
			s, err := json.Marshal(record)
			if err != nil {
				log.Printf("logIPFIX err=%v", err)
				continue
			}
			logCh <- &datastore.LogEnt{
				Time: time.Now().UnixNano(),
				Type: "ipfix",
				Log:  string(s),
			}
			if _, ok := record["sourceIPv4Address"]; ok {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("logIPFIX err=%v", r)
						for k, v := range record {
							log.Printf("%v=%v", k, v)
						}
					}
				}()
				if _, ok := record["sourceTransportPort"]; ok {
					report.ReportFlow(
						record["sourceIPv4Address"].(net.IP).String(),
						int(record["sourceTransportPort"].(uint16)),
						record["destinationIPv4Address"].(net.IP).String(),
						int(record["destinationTransportPort"].(uint16)),
						int(record["protocolIdentifier"].(uint8)),
						int64(record["packetDeltaCount"].(uint64)),
						int64(record["octetDeltaCount"].(uint64)),
						time.Now().UnixNano(),
					)
				} else if _, ok := record["icmpTypeCodeIPv4"]; ok {
					report.ReportFlow(
						record["sourceIPv4Address"].(net.IP).String(),
						0,
						record["destinationIPv4Address"].(net.IP).String(),
						int(record["icmpTypeCodeIPv4"].(uint16)),
						1,
						int64(record["packetDeltaCount"].(uint64)),
						int64(record["octetDeltaCount"].(uint64)),
						time.Now().UnixNano(),
					)
				}
			} else if _, ok := record["sourceIPv6Address"]; ok {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("logIPFIX err=%v", r)
						for k, v := range record {
							log.Printf("%v=%v", k, v)
						}
					}
				}()
				prot, ok := record["protocolIdentifier"]
				if ok {
					switch prot.(uint8) {
					case 6, 17:
						report.ReportFlow(
							record["sourceIPv6Address"].(net.IP).String(),
							int(record["sourceTransportPort"].(uint16)),
							record["destinationIPv6Address"].(net.IP).String(),
							int(record["destinationTransportPort"].(uint16)),
							int(record["protocolIdentifier"].(uint8)),
							int64(record["packetDeltaCount"].(uint64)),
							int64(record["octetDeltaCount"].(uint64)),
							time.Now().UnixNano(),
						)
					case 1:
						report.ReportFlow(
							record["sourceIPv6Address"].(net.IP).String(),
							0,
							record["destinationIPv6Address"].(net.IP).String(),
							int(record["icmpTypeCodeIPv6"].(uint16)),
							1,
							int64(record["packetDeltaCount"].(uint64)),
							int64(record["octetDeltaCount"].(uint64)),
							time.Now().UnixNano(),
						)
					}
				} else {
					log.Printf("unknown IPFIX record=%#v", record)
				}
			} else {
				log.Printf("unknown IPFIX record=%#v", record)
			}
		}
	}
	return r
}

func logNetflow(p *netflow5.Packet) {
	var record = make(map[string]interface{})
	for _, r := range p.Records {
		record["srcAddr"] = r.SrcAddr
		record["srcPort"] = r.SrcPort
		record["dstAddr"] = r.DstAddr
		record["dstPort"] = r.DstPort
		record["nextHop"] = r.NextHop
		record["bytes"] = r.Bytes
		record["packets"] = r.Packets
		record["first"] = r.First
		record["last"] = r.Last
		record["tcpflags"] = r.TCPFlags
		record["tcpflagsStr"] = read.TCPFlags(r.TCPFlags)
		record["protocol"] = r.Protocol
		record["protocolStr"] = read.Protocol(r.Protocol)
		record["tos"] = r.ToS
		record["srcAs"] = r.SrcAS
		record["dstAs"] = r.DstAS
		record["srcMask"] = r.SrcMask
		record["dstMask"] = r.DstMask
		s, err := json.Marshal(record)
		if err != nil {
			fmt.Println(err)
		}
		logCh <- &datastore.LogEnt{
			Time: time.Now().UnixNano(),
			Type: "netflow",
			Log:  string(s),
		}
		if v, ok := record["srcAddr"]; ok && v != nil {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("logNetflow err=%v", r)
					for k, v := range record {
						log.Printf("%v=%v", k, v)
					}
				}
			}()
			report.ReportFlow(
				record["srcAddr"].(net.IP).String(),
				int(record["srcPort"].(uint16)),
				record["dstAddr"].(net.IP).String(),
				int(record["dstPort"].(uint16)),
				int(record["protocol"].(uint8)),
				int64(r.Packets),
				int64(r.Bytes),
				time.Now().UnixNano(),
			)
		}
	}
}
