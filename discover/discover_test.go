package discover

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"net/http"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/ping"
)

func TestDiscover(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ping.Start(ctx, &sync.WaitGroup{}, "")
	defer cancel()
	time.Sleep(time.Second * 1)
	statikFS := http.Dir("../")
	td, err := os.MkdirTemp("", "twsnmpfc_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)
	datastore.Init(ctx, td, statikFS, &sync.WaitGroup{})
	datastore.MapConf.MapName = "Test123"
	if err := datastore.SaveMapConf(); err != nil {
		t.Fatal(err)
	}
	datastore.MapConf.Community = "public"
	datastore.DiscoverConf.StartIP = "192.168.1.1"
	datastore.DiscoverConf.EndIP = "192.168.1.2"
	datastore.DiscoverConf.Retry = 1
	datastore.DiscoverConf.Timeout = 2
	datastore.DiscoverConf.Active = true
	err = StartDiscover()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 15)
	t.Log("Done")
}
