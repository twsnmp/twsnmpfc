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

	err = OpenDB(indb)
	if err != nil {
		t.Fatal(err)
	}
	MapConf.MapName = "Test123"
	if err := SaveMapConfToDB(); err != nil {
		t.Fatal(err)
	}
	backdb, err := getTmpDBFile()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(backdb)
	DBStats.BackupFile = backdb
	DBStats.BackupConfigOnly = true
	err = BackupDB()
	if err != nil {
		t.Fatal(err)
	}
	CloseDB()
	MapConf.MapName = ""
	err = OpenDB(backdb)
	if err != nil {
		t.Fatal(err)
	}
	if MapConf.MapName != "Test123" {
		t.Errorf("Backup MapName = '%s'", MapConf.MapName)
	}
	CloseDB()
}
