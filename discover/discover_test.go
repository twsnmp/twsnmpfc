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
	p := ping.NewPing(ctx)
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
	ds := datastore.NewDataStore(td, statikFS)
	ds.MapConf.MapName = "Test123"
	if err := ds.SaveMapConfToDB(); err != nil {
		t.Fatal(err)
	}
	ds.MapConf.Community = "public"
	ds.DiscoverConf.StartIP = "192.168.1.1"
	ds.DiscoverConf.EndIP = "192.168.1.2"
	ds.DiscoverConf.Retry = 1
	ds.DiscoverConf.Timeout = 2
	d := NewDiscover(ds, p)
	err = d.StartDiscover()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 15)
	t.Log("Done")
}
