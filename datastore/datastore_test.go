package datastore

import (
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
	statikFS, err := fs.New()
	if err != nil {
		t.Fatal(err)
	}
	td, err := ioutil.TempDir("", "twsnmpfc_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)
	ds := NewDataStore(td, statikFS)
	ds.MapConf.MapName = "Test123"
	if err := ds.SaveMapConfToDB(); err != nil {
		t.Fatal(err)
	}
	backdb, err := getTmpDBFile()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(backdb)
	ds.DBStats.BackupFile = backdb
	ds.DBStats.BackupConfigOnly = true
	err = ds.BackupDB()
	if err != nil {
		t.Fatal(err)
	}
	ds.CloseDB()
	ds.MapConf.MapName = ""
	err = ds.OpenDB(backdb)
	if err != nil {
		t.Fatal(err)
	}
	if ds.MapConf.MapName != "Test123" {
		t.Errorf("Backup MapName = '%s'", ds.MapConf.MapName)
	}
	ds.CloseDB()
}
