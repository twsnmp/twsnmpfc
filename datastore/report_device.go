package datastore

type DeviceEnt struct {
	ID         string // MAC Addr
	Name       string
	IP         string
	NodeID     string
	Vendor     string
	Score      float64
	ValidScore bool
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

func GetDevice(id string) *DeviceEnt {
	if v, ok := devices.Load(id); ok {
		return v.(*DeviceEnt)
	}
	return nil
}

func AddDevice(d *DeviceEnt) {
	devices.Store(d.ID, d)
}

func ForEachDevices(f func(*DeviceEnt) bool) {
	devices.Range(func(k, v interface{}) bool {
		d := v.(*DeviceEnt)
		return f(d)
	})
}
