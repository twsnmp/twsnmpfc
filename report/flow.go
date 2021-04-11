package report

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

type flowReportEnt struct {
	Time    int64
	SrcIP   string
	SrcPort int
	DstIP   string
	DstPort int
	Prot    int
	Bytes   int64
}

func ReportFlow(src string, sp int, dst string, dp, prot int, pkts, bytes int64, t int64) {
	if prot == 6 &&
		pkts < int64(datastore.ReportConf.DropFlowThTCPPacket) {
		log.Printf("drop flow pkts=%d", pkts)
		return
	}
	flowReportCh <- &flowReportEnt{
		Time:    t,
		SrcIP:   src,
		SrcPort: sp,
		DstIP:   dst,
		DstPort: dp,
		Prot:    prot,
		Bytes:   bytes,
	}
}

func ResetFlowsScore() {
	datastore.ForEachFlows(func(f *datastore.FlowEnt) bool {
		f.Penalty = 0
		setFlowPenalty(f)
		f.UpdateTime = time.Now().UnixNano()
		return true
	})
	calcFlowScore()
}

func ResetServersScore() {
	datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
		s.Penalty = 0
		setServerPenalty(s)
		s.UpdateTime = time.Now().UnixNano()
		return true
	})
	calcServerScore()
}

// getFlowDir : クライアント、サーバー、サービスを決定するアルゴリズム
func getFlowDir(fr *flowReportEnt) (server, client, service string) {
	guc1 := datastore.IsGlobalUnicast(fr.SrcIP)
	guc2 := datastore.IsGlobalUnicast(fr.DstIP)
	if !guc1 && !guc2 {
		// 両方ユニキャストでない場合は含めない。
		return
	}
	s1, ok1 := datastore.GetServiceName(fr.Prot, fr.SrcPort)
	s2, ok2 := datastore.GetServiceName(fr.Prot, fr.DstPort)
	if ok1 {
		if ok2 {
			if fr.SrcPort < fr.DstPort || !guc1 {
				// ポート番号の小さい方を優先、または、マルチキャストはサーバーとする
				server = fr.SrcIP
				client = fr.DstIP
				service = s1
			} else if fr.SrcPort == fr.DstPort {
				id := fmt.Sprintf("%s:%s", fr.DstIP, fr.SrcIP)
				if datastore.GetFlow(id) != nil || !guc2 {
					// 既に登録済みか、マルチキャストをサーバーとする
					server = fr.DstIP
					client = fr.SrcIP
					service = s2
				} else {
					server = fr.SrcIP
					client = fr.DstIP
					service = s1
				}
			} else {
				server = fr.DstIP
				client = fr.SrcIP
				service = s2
			}
		} else {
			server = fr.SrcIP
			client = fr.DstIP
			service = s1
		}
	} else {
		if ok2 {
			server = fr.DstIP
			client = fr.SrcIP
			service = s2
		} else {
			if fr.SrcPort < fr.DstPort || !guc1 {
				server = fr.SrcIP
				client = fr.DstIP
				service = s1
			} else {
				server = fr.DstIP
				client = fr.SrcIP
				service = s2
			}
		}
	}
	return
}

func checkFlowReport(fr *flowReportEnt) {
	server, client, service := getFlowDir(fr)
	if server == "" {
		log.Printf("Skip flow report %v", fr)
		return
	}
	checkServerReport(server, service, fr.Bytes, fr.Time)
	now := time.Now().UnixNano()
	id := fmt.Sprintf("%s:%s", client, server)
	f := datastore.GetFlow(id)
	if f != nil {
		if _, ok := f.Services[service]; ok {
			f.Services[service]++
		} else {
			f.Services[service] = 1
			setFlowPenalty(f)
		}
		if f.ServerLoc == "" {
			f.ServerLoc = datastore.GetLoc(f.Server)
		}
		if f.ClientLoc == "" {
			f.ClientLoc = datastore.GetLoc(f.Client)
		}
		f.Bytes += fr.Bytes
		f.Count++
		f.LastTime = fr.Time
		f.UpdateTime = now
		return
	}
	f = &datastore.FlowEnt{
		ID:         id,
		Client:     client,
		Server:     server,
		Services:   make(map[string]int64),
		Count:      1,
		Bytes:      fr.Bytes,
		ServerLoc:  datastore.GetLoc(server),
		ClientLoc:  datastore.GetLoc(client),
		FirstTime:  fr.Time,
		LastTime:   fr.Time,
		UpdateTime: now,
	}
	f.ClientName, f.ClientNodeID = findNodeInfoFromIP(client)
	f.ServerName, f.ServerNodeID = findNodeInfoFromIP(server)
	f.Services[service] = 1
	setFlowPenalty(f)
	datastore.AddFlow(f)
}

func checkServerReport(server, service string, bytes, t int64) {
	if !strings.Contains(service, "/") {
		return
	}
	now := time.Now().UnixNano()
	id := server
	s := datastore.GetServer(id)
	if s != nil {
		if _, ok := s.Services[service]; ok {
			s.Services[service]++
		} else {
			s.Services[service] = 1
			setServerPenalty(s)
		}
		s.Count++
		s.Bytes += bytes
		s.LastTime = t
		s.UpdateTime = now
		s.ServerName, s.ServerNodeID = findNodeInfoFromIP(server)
		return
	}
	s = &datastore.ServerEnt{
		ID:         id,
		Server:     server,
		Services:   make(map[string]int64),
		Loc:        datastore.GetLoc(server),
		Count:      1,
		Bytes:      bytes,
		FirstTime:  t,
		LastTime:   t,
		UpdateTime: now,
	}
	s.ServerName, s.ServerNodeID = findNodeInfoFromIP(server)
	s.Services[service] = 1
	setServerPenalty(s)
	datastore.AddServer(s)
}

func setFlowPenalty(f *datastore.FlowEnt) {
	f.Penalty = 0
	if !isSafeCountry(f.ServerLoc) {
		f.Penalty++
	}
	for sv := range f.Services {
		if !isSafeService(sv, f.Server) {
			f.Penalty++
		}
	}
	// DNSで解決できない場合
	if f.ServerName == f.Server {
		f.Penalty++
	}
	if f.Penalty > 1 {
		if _, ok := badIPs[f.Client]; !ok {
			badIPs[f.Client] = true
		}
	}
}

func setServerPenalty(s *datastore.ServerEnt) {
	s.Penalty = 0
	if !isSafeCountry(s.Loc) {
		s.Penalty++
	}
	for sv := range s.Services {
		if !isSafeService(sv, s.Server) {
			s.Penalty++
		}
	}
	// DNSで解決できない場合
	if s.ServerName == s.Server {
		s.Penalty++
	}
}
