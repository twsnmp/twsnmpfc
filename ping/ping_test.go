package ping

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	Start(ctx, wg, "")
	defer cancel()
	time.Sleep(time.Second * 1)
	ip := ""
	ok := false
	if ip, ok = os.LookupEnv("OK_IP"); !ok {
		ip = "192.168.1.1"
	}
	r := DoPing(ip, 1, 1, 12, 0)
	if r.Stat != PingOK {
		t.Errorf("ping stat = %d", r.Stat)
	}
	if ip, ok = os.LookupEnv("NG_IP"); !ok {
		ip = "192.168.1.33"
	}
	r = DoPing(ip, 1, 1, 12, 0)
	if r.Stat == PingOK {
		t.Errorf("ping stat = %d", r.Stat)
	}
	t.Log("Done")
}
