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
		f.ClientName, f.ClientNodeID = findNodeInfoFromIP(f.Client)
		f.ServerName, f.ServerNodeID = findNodeInfoFromIP(f.Server)
		setFlowPenalty(f)
		f.UpdateTime = time.Now().UnixNano()
		return true
	})
	calcFlowScore()
}

func ResetServersScore() {
	datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
		s.Penalty = 0
		s.ServerName, s.ServerNodeID = findNodeInfoFromIP(s.Server)
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
	if !guc1 && !guc2 && !strings.HasPrefix(fr.SrcIP, "fe80::") &&
		!strings.HasPrefix(fr.SrcIP, "169.254.") {
		// 両方ユニキャストでない場合は含めない。IPv4,IPv6のリンクローカルは含める
		return
	}
	if fr.Prot == 1 {
		// ICMP
		return getFlowDirICMP(fr)
	}
	if fr.Prot == 2 {
		server = fr.SrcIP
		client = fr.DstIP
		service = "igmp"
		return
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
	} else {
		// サービスが不明の場合で他のプロトコルで登録済みの場合
		if datastore.GetFlow(fmt.Sprintf("%s:%s", fr.SrcIP, fr.DstIP)) != nil {
			server = fr.DstIP
			client = fr.SrcIP
			service = getOtherProtName(fr.Prot)
		} else if datastore.GetFlow(fmt.Sprintf("%s:%s", fr.DstIP, fr.SrcIP)) != nil {
			server = fr.SrcIP
			client = fr.DstIP
			service = getOtherProtName(fr.Prot)
		} else {
			server = fr.DstIP
			client = fr.SrcIP
			service = getOtherProtName(fr.Prot)
			addToUnknownPortMap(fr)
		}
	}
	return
}

func getOtherProtName(prot int) string {
	if prot == 6 {
		return "other/tcp"
	}
	if prot == 17 {
		return "other/udp"
	}
	return fmt.Sprintf("other/%d", prot)
}

var UnKnownPortMap = make(map[string]int64)

func addToUnknownPortMap(fr *flowReportEnt) {
	var skey string
	var dkey string
	switch fr.Prot {
	case 6:
		skey = fmt.Sprintf("%d/tcp", fr.SrcPort)
		dkey = fmt.Sprintf("%d/tcp", fr.DstPort)
	case 17:
		skey = fmt.Sprintf("%d/udp", fr.SrcPort)
		dkey = fmt.Sprintf("%d/udp", fr.DstPort)
	default:
		return
	}
	if _, ok := UnKnownPortMap[skey]; !ok {
		UnKnownPortMap[skey] = 1
	} else {
		UnKnownPortMap[skey]++
	}
	if _, ok := UnKnownPortMap[dkey]; !ok {
		UnKnownPortMap[dkey] = 1
	} else {
		UnKnownPortMap[dkey]++
	}
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
	for k, fr := range udpPending {
		if fr.Time < time.Now().UnixNano()-(1000*1000*1000*10) {
			delete(udpPending, k)
		}
	}
}

func checkFlowReport(fr *flowReportEnt) {
	checkIPReport(fr.SrcIP, "", fr.Time)
	server, client, service := getFlowDir(fr)
	if server == "" {
		return
	}
	if fr.Prot == 17 && datastore.IsGlobalUnicast(fr.DstIP) {
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

func checkOldServers(safeOld, delOld int64) {
	ids := []string{}
	datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
		if s.LastTime < safeOld {
			if s.LastTime < delOld || s.Score > 50.0 || s.Count < 10 {
				ids = append(ids, s.ID)
			}
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("servers", ids)
		log.Printf("report delete servers=%d", len(ids))
	}
}

func checkOldFlows(safeOld, delOld int64) {
	ids := []string{}
	datastore.ForEachFlows(func(f *datastore.FlowEnt) bool {
		if f.LastTime < safeOld {
			if f.LastTime < delOld || f.Score > 50.0 || f.Count < 10 {
				ids = append(ids, f.ID)
			}
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("flows", ids)
		log.Printf("report delete flows=%d", len(ids))
	}
}

func calcServerScore() {
	var xs []float64
	datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
		if s.Penalty > 100 {
			s.Penalty = 100
		}
		xs = append(xs, float64(100-s.Penalty))
		return true
	})
	m, sd := getMeanSD(&xs)
	datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
		if sd != 0 {
			s.Score = ((10 * (float64(100-s.Penalty) - m) / sd) + 50)
		} else {
			s.Score = 50.0
		}
		s.ValidScore = true
		return true
	})
}

func calcFlowScore() {
	var xs []float64
	datastore.ForEachFlows(func(f *datastore.FlowEnt) bool {
		if f.Penalty > 100 {
			f.Penalty = 100
		}
		xs = append(xs, float64(100-f.Penalty))
		return true
	})
	m, sd := getMeanSD(&xs)
	datastore.ForEachFlows(func(f *datastore.FlowEnt) bool {
		if sd != 0 {
			f.Score = ((10 * (float64(100-f.Penalty) - m) / sd) + 50)
		} else {
			f.Score = 50.0
		}
		f.ValidScore = true
		return true
	})
}
