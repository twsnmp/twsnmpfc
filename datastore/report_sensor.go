package datastore

type SensorEnt struct {
	ID        string // Host + Type + Param
	Host      string
	Type      string // twpcap,twwinlog....
	Param     string
	Total     int64
	Send      int64
	State     string
	Stats     []SensorStatsEnt
	Monitors  []SensorMonitorEnt
	FirstTime int64
	LastTime  int64
}

type SensorStatsEnt struct {
	Time     int64
	Total    int64
	Count    int64
	PS       float64
	Send     int64
	LastSend int64
}

type SensorMonitorEnt struct {
	Time    int64
	CPU     float64
	Mem     float64
	Load    float64
	Process float64
	Recv    int64
	Sent    int64
	TxSpeed float64
	RxSpeed float64
}

func GetSensor(id string) *SensorEnt {
	if v, ok := sensors.Load(id); ok {
		return v.(*SensorEnt)
	}
	return nil
}

func AddSensor(s *SensorEnt) {
	sensors.Store(s.ID, s)
}

func ForEachSensors(f func(*SensorEnt) bool) {
	sensors.Range(func(k, v interface{}) bool {
		s := v.(*SensorEnt)
		return f(s)
	})
}
