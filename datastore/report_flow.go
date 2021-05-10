package datastore

type ServerEnt struct {
	ID           string //  ID Server
	Server       string
	Services     map[string]int64
	Count        int64
	Bytes        int64
	ServerName   string
	ServerNodeID string
	Loc          string
	Score        float64
	ValidScore   bool
	Penalty      int64
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

type FlowEnt struct {
	ID           string // ID Client:Server
	Client       string
	Server       string
	Services     map[string]int64
	Count        int64
	Bytes        int64
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

func GetFlow(id string) *FlowEnt {
	if v, ok := flows.Load(id); ok {
		return v.(*FlowEnt)
	}
	return nil
}

func AddFlow(f *FlowEnt) {
	flows.Store(f.ID, f)
}

func ForEachFlows(f func(*FlowEnt) bool) {
	flows.Range(func(k, v interface{}) bool {
		fl := v.(*FlowEnt)
		return f(fl)
	})
}

func GetServer(id string) *ServerEnt {
	if v, ok := servers.Load(id); ok {
		return v.(*ServerEnt)
	}
	return nil
}

func AddServer(s *ServerEnt) {
	servers.Store(s.ID, s)
}

func ForEachServers(f func(*ServerEnt) bool) {
	servers.Range(func(k, v interface{}) bool {
		s := v.(*ServerEnt)
		return f(s)
	})
}
