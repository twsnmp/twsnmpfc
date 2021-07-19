package datastore

type IPReportEnt struct {
	IP         string
	MAC        string
	Name       string
	NodeID     string
	Loc        string
	Vendor     string
	Count      int64
	Change     int64
	Score      float64
	ValidScore bool
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

func GetIPReport(id string) *IPReportEnt {
	if v, ok := ips.Load(id); ok {
		return v.(*IPReportEnt)
	}
	return nil
}

func AddIPReport(ip *IPReportEnt) {
	ips.Store(ip.IP, ip)
}

func ForEachIPReport(f func(*IPReportEnt) bool) {
	ips.Range(func(k, v interface{}) bool {
		ip := v.(*IPReportEnt)
		return f(ip)
	})
}

func findIPsFromMAC(mac string) []string {
	var ret = []string{}
	ips.Range(func(k, v interface{}) bool {
		ip := v.(*IPReportEnt)
		if ip.MAC == mac {
			ret = append(ret, ip.IP)
		}
		return true
	})
	return ret
}
