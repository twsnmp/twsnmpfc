package logger

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/Cistern/sflow"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/tehmaze/netflow/read"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
)

func sflowd(stopCh chan bool) {
	log.Printf("start sflowd")
	sv, err := net.ListenPacket("udp", sflowListen)
	if err != nil {
		log.Printf("sflowd err=%v", err)
		<-stopCh
		return
	}
	defer sv.Close()
	data := make([]byte, 8192)
	for {
		select {
		case <-stopCh:
			log.Printf("stop sflowd")
			return
		default:
			l, ra, err := sv.ReadFrom(data)
			if err != nil {
				continue
			}
			r := bytes.NewReader(data[:l])
			d := sflow.NewDecoder(r)
			dg, err := d.Decode()
			if err != nil {
				log.Printf("sflow decode err=%v", err)
				continue
			}
			raIP := ""
			switch a := ra.(type) {
			case *net.UDPAddr:
				raIP = a.IP.String()
			case *net.TCPAddr:
				raIP = a.IP.String()
			}
			report.UpdateSensor(raIP, "sflow", len(dg.Samples))
			for _, sample := range dg.Samples {
				switch s := sample.(type) {
				case *sflow.CounterSample:
					for _, record := range s.Records {
						switch csr := record.(type) {
						case sflow.HostDiskCounters:
							s, err := json.Marshal(record)
							if err != nil {
								log.Println(err)
								continue
							}
							sFlowCounter("HostDiskCounter", raIP, string(s))
						case sflow.HostCPUCounters:
							s, err := json.Marshal(record)
							if err != nil {
								log.Println(err)
								continue
							}
							sFlowCounter("HostCPUCounter", raIP, string(s))
						case sflow.HostMemoryCounters:
							s, err := json.Marshal(record)
							if err != nil {
								log.Println(err)
								continue
							}
							sFlowCounter("HostMemoryCounter", raIP, string(s))
						case sflow.HostNetCounters:
							s, err := json.Marshal(record)
							if err != nil {
								log.Println(err)
								continue
							}
							sFlowCounter("HostNetCounter", raIP, string(s))
						case sflow.GenericInterfaceCounters:
							s, err := json.Marshal(record)
							if err != nil {
								log.Println(err)
								continue
							}
							sFlowCounter("GenericInterfaceCounter", raIP, string(s))
						default:
							log.Printf("sflow unknown counter sample %v", csr)
							continue
						}
					}
				case *sflow.FlowSample:
					for _, record := range s.Records {
						switch fsr := record.(type) {
						case sflow.RawPacketFlow:
							rawPacketFlow(&fsr, 0)
						}
					}
				case *sflow.EventDiscardedPacket:
					for _, record := range s.Records {
						switch fsr := record.(type) {
						case sflow.RawPacketFlow:
							rawPacketFlow(&fsr, int(s.Reason))
						}
					}
				}
			}
		}
	}
}

func rawPacketFlow(r *sflow.RawPacketFlow, reason int) {
	packet := gopacket.NewPacket(r.Header, layers.LayerTypeEthernet, gopacket.Default)
	var record = make(map[string]interface{})
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethernetLayer != nil {
		if eth, ok := ethernetLayer.(*layers.Ethernet); ok {
			record["sourceMacAddress"] = eth.SrcMAC.String()
			record["destinationMacAddress"] = eth.SrcMAC.String()
		}
	}
	src := ""
	dst := ""
	bytes := 0
	ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
	if ipv4Layer != nil {
		ip, ok := ipv4Layer.(*layers.IPv4)
		if !ok {
			return
		}
		src = ip.SrcIP.String()
		dst = ip.DstIP.String()
		bytes = int(ip.Length)
	} else {
		ipv6Layer := packet.Layer(layers.LayerTypeIPv6)
		if ipv6Layer != nil {
			ipv6, ok := ipv6Layer.(*layers.IPv6)
			if !ok {
				return
			}
			src = ipv6.SrcIP.String()
			dst = ipv6.DstIP.String()
			bytes = int(ipv6.Length)
		}
	}
	sp := 0
	dp := 0
	prot := 0
	flag := uint8(0)
	// UDP
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer != nil {
		udp, ok := udpLayer.(*layers.UDP)
		if !ok {
			return
		}
		sp = int(udp.SrcPort)
		dp = int(udp.DstPort)
		prot = 17
	} else {
		// TCP
		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer != nil {
			tcp, ok := tcpLayer.(*layers.TCP)
			if !ok {
				return
			}
			sp = int(tcp.SrcPort)
			dp = int(tcp.DstPort)
			if tcp.FIN {
				flag |= 0x01
			}
			if tcp.SYN {
				flag |= 0x02
			}
			if tcp.RST {
				flag |= 0x04
			}
			if tcp.PSH {
				flag |= 0x08
			}
			if tcp.ACK {
				flag |= 0x10
			}
			if tcp.URG {
				flag |= 0x20
			}
			if tcp.ECE {
				flag |= 0x40
			}
			if tcp.CWR {
				flag |= 0x80
			}
			prot = 6
		} else {
			icmpV4Layer := packet.Layer(layers.LayerTypeICMPv4)
			if icmpV4Layer != nil {
				icmp, ok := icmpV4Layer.(*layers.ICMPv4)
				if !ok {
					return
				}
				prot = 1
				dp = int(icmp.TypeCode)
			} else {
				icmpV6Layer := packet.Layer(layers.LayerTypeICMPv6)
				if icmpV6Layer == nil {
					return
				}
				icmp, ok := icmpV6Layer.(*layers.ICMPv6)
				if !ok {
					return
				}
				prot = 58
				dp = int(icmp.TypeCode)
			}
		}
	}
	record["srcAddr"] = src
	record["srcPort"] = float64(sp)
	record["dstAddr"] = dst
	record["dstPort"] = float64(dp)
	record["bytes"] = float64(bytes)
	record["packets"] = float64(1)
	record["tcpflags"] = float64(flag)
	record["tcpflagsStr"] = read.TCPFlags(flag)
	record["protocol"] = float64(prot)
	record["first"] = float64(0)
	record["last"] = float64(0)
	record["protocolStr"] = read.Protocol(uint8(prot))
	record["discardedReason"] = reason
	s, err := json.Marshal(record)
	if err != nil {
		log.Println(err)
		return
	}
	logCh <- &datastore.LogEnt{
		Time: time.Now().UnixNano(),
		Type: "sflow",
		Log:  string(s),
	}
	report.ReportFlow(
		src,
		sp,
		dst,
		dp,
		prot,
		1,
		int64(bytes),
		time.Now().UnixNano(),
	)
}

func sFlowCounter(t, r, d string) {
	record := datastore.SFlowCounterEnt{
		Type:   t,
		Remote: r,
		Data:   d,
	}
	s, err := json.Marshal(record)
	if err != nil {
		log.Println(err)
		return
	}
	logCh <- &datastore.LogEnt{
		Time: time.Now().UnixNano(),
		Type: "sflowCounter",
		Log:  string(s),
	}
}
