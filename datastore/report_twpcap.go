package datastore

type EtherTypeEnt struct {
	ID        string // ID Host:EtherType
	Host      string
	Type      string
	Name      string
	Count     int64
	FirstTime int64
	LastTime  int64
}

type DNSQEnt struct {
	ID           string // ID Host:Server:Type:Name
	Host         string
	Server       string
	Type         string
	Name         string
	Count        int64
	Change       int64
	ServerName   string
	ServerNodeID string
	ServerLoc    string
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

type RADIUSFlowEnt struct {
	ID           string // ID Client:Server
	Client       string
	Server       string
	Count        int64
	Request      int64
	Challenge    int64
	Accept       int64
	Reject       int64
	ClientName   string
	ClientNodeID string
	ServerName   string
	ServerNodeID string
	Score        float64
	ValidScore   bool
	Penalty      int64
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

type TLSFlowEnt struct {
	ID           string // ID Client:Server:Service
	Client       string
	Server       string
	Service      string
	Count        int64
	Version      string
	Cipher       string
	ClientName   string
	ClientNodeID string
	ClientLoc    string
	ServerName   string
	ServerNodeID string
	ServerLoc    string
	Score        float64
	ValidScore   bool
	Penalty      int64
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

func GetEtherType(id string) *EtherTypeEnt {
	if v, ok := etherType.Load(id); ok {
		return v.(*EtherTypeEnt)
	}
	return nil
}

func AddEtherType(s *EtherTypeEnt) {
	etherType.Store(s.ID, s)
}

func ForEachEtherType(f func(*EtherTypeEnt) bool) {
	etherType.Range(func(k, v interface{}) bool {
		s := v.(*EtherTypeEnt)
		return f(s)
	})
}

func GetDNSQ(id string) *DNSQEnt {
	if v, ok := dnsq.Load(id); ok {
		return v.(*DNSQEnt)
	}
	return nil
}

func AddDNSQ(s *DNSQEnt) {
	dnsq.Store(s.ID, s)
}

func ForEachDNSQ(f func(*DNSQEnt) bool) {
	dnsq.Range(func(k, v interface{}) bool {
		s := v.(*DNSQEnt)
		return f(s)
	})
}

func GetRADIUSFlow(id string) *RADIUSFlowEnt {
	if v, ok := radiusFlows.Load(id); ok {
		return v.(*RADIUSFlowEnt)
	}
	return nil
}

func AddRADIUSFlow(f *RADIUSFlowEnt) {
	radiusFlows.Store(f.ID, f)
}

func ForEachRADIUSFlows(f func(*RADIUSFlowEnt) bool) {
	radiusFlows.Range(func(k, v interface{}) bool {
		fl := v.(*RADIUSFlowEnt)
		return f(fl)
	})
}

func GetTLSFlow(id string) *TLSFlowEnt {
	if v, ok := tlsFlows.Load(id); ok {
		return v.(*TLSFlowEnt)
	}
	return nil
}

func AddTLSFlow(f *TLSFlowEnt) {
	tlsFlows.Store(f.ID, f)
}

func ForEachTLSFlows(f func(*TLSFlowEnt) bool) {
	tlsFlows.Range(func(k, v interface{}) bool {
		fl := v.(*TLSFlowEnt)
		return f(fl)
	})
}
