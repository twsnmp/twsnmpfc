package discover

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/rakyll/statik/fs"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/ping"
	_ "github.com/twsnmp/twsnmpfc/statik"
)

func getTmpDBFile() (string, error) {
	f, err := ioutil.TempFile("", "twsnmpfc_test")
	if err != nil {
		return "", err
	}
	return f.Name(), err
}

func TestDiscover(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ping.Start(ctx)
	defer cancel()
	time.Sleep(time.Second * 1)
	statikFS, err := fs.New()
	if err != nil {
		t.Fatal(err)
	}
	td, err := ioutil.TempDir("", "twsnmpfc_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)
	datastore.Init(ctx, td, statikFS)
	datastore.MapConf.MapName = "Test123"
	if err := datastore.SaveMapConf(); err != nil {
		t.Fatal(err)
	}
	datastore.MapConf.Community = "public"
	datastore.DiscoverConf.StartIP = "192.168.1.1"
	datastore.DiscoverConf.EndIP = "192.168.1.2"
	datastore.DiscoverConf.Retry = 1
	datastore.DiscoverConf.Timeout = 2
	err = StartDiscover()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 15)
	t.Log("Done")
}
