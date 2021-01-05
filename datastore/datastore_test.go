package datastore

import (
	"io/ioutil"
	"os"
	"testing"
)

func getTmpDBFile() (string, error) {
	f, err := ioutil.TempFile("", "twsnmpfc_test")
	if err != nil {
		return "", err
	}
	return f.Name(), err
}

func TestDataStore(t *testing.T) {
	indb, err := getTmpDBFile()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(indb)
	ds := NewDataStore()
	err = ds.OpenDB(indb)
	if err != nil {
		t.Fatal(err)
	}
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
