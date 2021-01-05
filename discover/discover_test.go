package discover

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/rakyll/statik/fs"
	mibdb "github.com/twsnmp/go-mibdb"
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
	dbf, err := getTmpDBFile()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dbf)
	ds := datastore.NewDataStore()
	err = ds.OpenDB(dbf)
	if err != nil {
		t.Fatal(err)
	}
	ds.MapConf.MapName = "Test123"
	if err := ds.SaveMapConfToDB(); err != nil {
		t.Fatal(err)
	}
	statikFS, err := fs.New()
	if err != nil {
		t.Fatal(err)
	}
	r, err := statikFS.Open("/conf/mib.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	s, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	mib, err := mibdb.NewMIBDBFromStr(string(s), "")
	if err != nil {
		t.Fatal(err)
	}
	ds.MapConf.Community = "public"
	ds.DiscoverConf.StartIP = "192.168.1.1"
	ds.DiscoverConf.EndIP = "192.168.1.2"
	ds.DiscoverConf.Retry = 1
	ds.DiscoverConf.Timeout = 2
	ds.DiscoverConf.SnmpMode = "snmpv1"
	d, err := NewDiscover(ds, p, mib)
	if err != nil {
		t.Fatal(err)
	}
	err = d.StartDiscover()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Done")
}
