package datastore

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/rakyll/statik/fs"
	_ "github.com/twsnmp/twsnmpfc/statik"
)

func getTmpDBFile() (string, error) {
	f, err := ioutil.TempFile("", "twsnmpfc_test")
	if err != nil {
		return "", err
	}
	return f.Name(), err
}

func TestDataStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	statikFS, err := fs.New()
	if err != nil {
		t.Fatal(err)
	}
	td, err := ioutil.TempDir("", "twsnmpfc_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)
	InitDataStore(ctx, td, statikFS)
	MapConf.MapName = "Test123"
	if err := SaveMapConfToDB(); err != nil {
		t.Fatal(err)
	}
	defer cancel()
	backdb, err := getTmpDBFile()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(backdb)
	Backup.ConfigOnly = true
	err = BackupDB(backdb)
	if err != nil {
		t.Fatal(err)
	}
	CloseDataStore()
	MapConf.MapName = ""
	err = openDB(backdb)
	if err != nil {
		t.Fatal(err)
	}
	if MapConf.MapName != "Test123" {
		t.Errorf("Backup MapName = '%s'", MapConf.MapName)
	}
	CloseDataStore()
}
