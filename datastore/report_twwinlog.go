package datastore

type WinEventIDEnt struct {
	ID        string // Sender + Computer + Provider + EventID
	Sender    string
	Computer  string
	Provider  string
	EventID   int
	Count     int64
	FirstTime int64
	LastTime  int64
}

func GetWinEventID(id string) *WinEventIDEnt {
	if v, ok := wineventIDs.Load(id); ok {
		return v.(*WinEventIDEnt)
	}
	return nil
}

func AddWinEventID(s *WinEventIDEnt) {
	wineventIDs.Store(s.ID, s)
}

func ForEachWinEventID(f func(*WinEventIDEnt) bool) {
	wineventIDs.Range(func(k, v interface{}) bool {
		s := v.(*WinEventIDEnt)
		return f(s)
	})
}
