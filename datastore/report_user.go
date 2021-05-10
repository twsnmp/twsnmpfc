package datastore

type UserClientEnt struct {
	Total int32
	Ok    int32
}

type UserEnt struct {
	ID           string // User ID + Server
	UserID       string
	Server       string
	ServerName   string
	ServerNodeID string
	ClientMap    map[string]UserClientEnt
	Total        int
	Ok           int
	Score        float64
	ValidScore   bool
	Penalty      int64
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

func GetUser(id string) *UserEnt {
	if v, ok := users.Load(id); ok {
		return v.(*UserEnt)
	}
	return nil
}

func AddUser(u *UserEnt) {
	users.Store(u.ID, u)
}

func ForEachUsers(f func(*UserEnt) bool) {
	users.Range(func(k, v interface{}) bool {
		u := v.(*UserEnt)
		return f(u)
	})
}

func DeleteUser(id string) {
	users.Delete(id)
}
