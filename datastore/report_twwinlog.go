package datastore

type WinEventIDEnt struct {
	ID        string // Computer + Provider + EventID
	Level     string
	Computer  string
	Provider  string
	Channel   string
	EventID   int
	Count     int64
	FirstTime int64
	LastTime  int64
}

func GetWinEventID(id string) *WinEventIDEnt {
	if v, ok := winEventID.Load(id); ok {
		return v.(*WinEventIDEnt)
	}
	return nil
}

func AddWinEventID(s *WinEventIDEnt) {
	winEventID.Store(s.ID, s)
}

func ForEachWinEventID(f func(*WinEventIDEnt) bool) {
	winEventID.Range(func(k, v interface{}) bool {
		s := v.(*WinEventIDEnt)
		return f(s)
	})
}
