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

func (r *Report) ReportFlow(src string, sp int, dst string, dp, prot int, bytes int64, t int64) {
	r.flowReportCh <- &flowReportEnt{
		Time:    t,
		SrcIP:   src,
		SrcPort: sp,
		DstIP:   dst,
		DstPort: dp,
		Prot:    prot,
		Bytes:   bytes,
	}
}

// getFlowDir : クライアント、サーバー、サービスを決定するアルゴリズム
func (r *Report) getFlowDir(fr *flowReportEnt) (server, client, service string) {
	guc1 := datastore.IsGlobalUnicast(fr.SrcIP)
	guc2 := datastore.IsGlobalUnicast(fr.DstIP)
	if !guc1 && !guc2 {
		// 両方ユニキャストでない場合は含めない。
		return
	}
	s1, ok1 := r.ds.GetServiceName(fr.Prot, fr.SrcPort)
	s2, ok2 := r.ds.GetServiceName(fr.Prot, fr.DstPort)
	if ok1 {
		if ok2 {
			if fr.SrcPort < fr.DstPort || !guc1 {
				// ポート番号の小さい方を優先、または、マルチキャストはサーバーとする
				server = fr.SrcIP
				client = fr.DstIP
				service = s1
			} else if fr.SrcPort == fr.DstPort {
				id := fmt.Sprintf("%s:%s", fr.DstIP, fr.SrcIP)
				if r.ds.GetFlow(id) != nil || !guc2 {
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

func (r *Report) checkFlowReport(fr *flowReportEnt) {
	server, client, service := r.getFlowDir(fr)
	if server == "" {
		log.Printf("Skip flow report %v", r)
		return
	}
	r.checkServerReport(server, service, fr.Bytes, fr.Time)
	now := time.Now().UnixNano()
	id := fmt.Sprintf("%s:%s", client, server)
	f := r.ds.GetFlow(id)
	if f != nil {
		if _, ok := f.Services[service]; ok {
			f.Services[service]++
		} else {
			f.Services[service] = 1
			r.setFlowPenalty(f)
		}
		if f.ServerLoc == "" {
			f.ServerLoc = r.ds.GetLoc(f.Server)
		}
		if f.ClientLoc == "" {
			f.ClientLoc = r.ds.GetLoc(f.Client)
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
		ServerLoc:  r.ds.GetLoc(server),
		ClientLoc:  r.ds.GetLoc(client),
		ServerName: r.findNameFromIP(server),
		ClientName: r.findNameFromIP(client),
		FirstTime:  fr.Time,
		LastTime:   fr.Time,
		UpdateTime: now,
	}
	f.Services[service] = 1
	r.setFlowPenalty(f)
	r.ds.AddFlow(f)
}

func (r *Report) checkServerReport(server, service string, bytes, t int64) {
	if !strings.Contains(service, "/") {
		return
	}
	now := time.Now().UnixNano()
	id := server
	s := r.ds.GetServer(id)
	if s != nil {
		if _, ok := s.Services[service]; ok {
			s.Services[service]++
		} else {
			s.Services[service] = 1
			r.setServerPenalty(s)
		}
		s.Count++
		s.Bytes += bytes
		s.LastTime = t
		s.UpdateTime = now
		return
	}
	s = &datastore.ServerEnt{
		ID:         id,
		Server:     server,
		Services:   make(map[string]int64),
		ServerName: r.findNameFromIP(server),
		Loc:        r.ds.GetLoc(server),
		Count:      1,
		Bytes:      bytes,
		FirstTime:  t,
		LastTime:   t,
		UpdateTime: now,
	}
	s.Services[service] = 1
	r.setServerPenalty(s)
	r.ds.AddServer(s)
}

func (r *Report) setFlowPenalty(f *datastore.FlowEnt) {
	loc := ""
	if f.ServerLoc != "" {
		a := strings.Split(f.ServerLoc, ",")
		if len(a) > 0 {
			loc = a[0]
		}
	}
	f.Penalty = 0
	ids := []string{}
	for service := range f.Services {
		ids = append(ids, fmt.Sprintf("*:%s:*", service))
		if loc != "" {
			ids = append(ids, fmt.Sprintf("*:%s:%s", service, loc))
		}
		if as := r.ds.GetAllowRule(service); as != nil {
			if e, ok := as.Servers[f.Server]; !ok {
				if e {
					f.Penalty++
				}
			}
		}
	}
	ids = append(ids, fmt.Sprintf("%s:*:*", f.Server))
	if loc != "" {
		ids = append(ids, fmt.Sprintf("*:*:%s", loc))
	}
	for _, id := range ids {
		if r.ds.GetDennyRule(id) {
			f.Penalty++
		}
	}
	// DNSで解決できない場合
	if f.ServerName == f.Server {
		f.Penalty++
	}
	if f.Penalty > 0 {
		if n, ok := r.badIPs[f.Client]; !ok || n < f.Penalty {
			r.badIPs[f.Client] = f.Penalty
		}
	}
}

func (r *Report) setServerPenalty(s *datastore.ServerEnt) {
	loc := ""
	if s.Loc != "" {
		a := strings.Split(s.Loc, ",")
		if len(a) > 0 {
			loc = a[0]
		}
	}
	s.Penalty = 0
	ids := []string{}
	for service := range s.Services {
		ids = append(ids, fmt.Sprintf("*:%s:*", service))
		if loc != "" {
			ids = append(ids, fmt.Sprintf("*:%s:%s", service, loc))
		}
		if as := r.ds.GetAllowRule(service); as != nil {
			if _, ok := as.Servers[s.Server]; !ok {
				s.Penalty++
			}
		}
	}
	if loc != "" {
		ids = append(ids, fmt.Sprintf("*:*:%s", loc))
	}
	for _, id := range ids {
		if r.ds.GetDennyRule(id) {
			s.Penalty++
		}
	}
	// DNSで解決できない場合
	if s.ServerName == s.Server {
		s.Penalty++
	}
}
