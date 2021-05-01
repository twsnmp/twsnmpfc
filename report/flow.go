package report

import (
	"fmt"
	"log"
	"net"
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
	if !guc1 && !guc2 && !strings.HasPrefix(fr.SrcIP, "fe80::") {
		// 両方ユニキャストでない場合は含めない。IPv6のリンクローカルは含める
		log.Println("getFlowDir src and dst address is not unicast")
		return
	}
	if fr.Prot == 1 {
		// ICMP
		return getFlowDirICMP(fr)
	}
	s1, ok1 := datastore.GetServiceName(fr.Prot, fr.SrcPort)
	s2, ok2 := datastore.GetServiceName(fr.Prot, fr.DstPort)
	if ok1 {
		if ok2 {
			// 両方サービス名がわかる時は
			if !guc2 || IsDstServer(fr) {
				// マルチキャスト、アドレスの関係で判断してサーバーを決める
				server = fr.DstIP
				client = fr.SrcIP
				service = s2
			} else {
				server = fr.SrcIP
				client = fr.DstIP
				service = s1
			}
		} else {
			server = fr.SrcIP
			client = fr.DstIP
			service = s1
		}
	} else if ok2 {
		server = fr.DstIP
		client = fr.SrcIP
		service = s2
	}
	//サービス名が不明は捨てる
	return
}

// IsDstServer : Dstがサーバーならばtrueを返す
func IsDstServer(fr *flowReportEnt) bool {
	// サブネットマスクがわからないが、ブロードキャストぽいもので判定
	if strings.HasSuffix(fr.DstIP, ".255") {
		return true
	}
	if strings.HasSuffix(fr.SrcIP, ".255") {
		return false
	}
	srcP := isPrivateAddr(fr.SrcIP)
	dstP := isPrivateAddr(fr.DstIP)
	if srcP && !dstP {
		return true
	}
	if !srcP && dstP {
		return false
	}
	// ポート番号の小さいほうをサーバーにする
	return fr.SrcPort >= fr.DstPort
}

var privateAddrList = []*net.IPNet{}
var privateAddrStrList = []string{
	"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16 ",
}

func isPrivateAddr(a string) bool {
	ip := net.ParseIP(a)
	// Priavate Addressではない方
	if len(privateAddrList) < 1 {
		for _, ps := range privateAddrStrList {
			if _, cidr, err := net.ParseCIDR(ps); err == nil {
				privateAddrList = append(privateAddrList, cidr)
			}
		}
	}
	for _, cidr := range privateAddrList {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func getFlowDirICMP(fr *flowReportEnt) (server, client, service string) {
	icmpType := (fr.DstPort / 256)
	switch icmpType {
	case 0, 3, 11, 12:
		// Rourer Serverが応答するもの
		service = fmt.Sprintf("%d/icmp", icmpType)
		server = fr.SrcIP
		client = fr.DstIP
	case 5, 8, 13:
		// Clientから送信するもの
		service = fmt.Sprintf("%d/icmp", icmpType)
		server = fr.DstIP
		client = fr.SrcIP
	default:
		//未定義廃止されたicmp
		service = "-1/icmp"
		server = fr.DstIP
		client = fr.SrcIP
	}
	return
}

var udpPending = make(map[string]*flowReportEnt)

// 10秒以内に逆方法の通信がないケースはレポート対象外
func cleanupUDPPending() {
	count := 0
	for k, fr := range udpPending {
		if fr.Time < time.Now().UnixNano()-(1000*1000*1000*10) {
			count++
			delete(udpPending, k)
		}
	}
	if count > 0 {
		log.Printf("delete pending udp flow count=%d", count)
	}
}

func checkFlowReport(fr *flowReportEnt) {
	server, client, service := getFlowDir(fr)
	if server == "" {
		log.Printf("skip flow report %v", fr)
		return
	}
	if fr.Prot == 17 {
		// UDP
		cleanupUDPPending()
		id := fmt.Sprintf("%s:%s:%s", client, server, service)
		if ufr, ok := udpPending[id]; ok {
			if ufr.DstIP == fr.SrcIP {
				//逆方向の通信がある場合だけ登録する
				fr.Bytes += ufr.Bytes
			} else {
				ufr.Bytes += fr.Bytes
				return
			}
		} else {
			udpPending[id] = fr
			return
		}
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
