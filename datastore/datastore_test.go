package datastore

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/rakyll/statik/fs"
	_ "github.com/twsnmp/twsnmpfc/statik"
)

func getTmpDBFile() (string, error) {
	f, err := os.CreateTemp("", "twsnmpfc_test")
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
	td, err := os.MkdirTemp("", "twsnmpfc_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)
	wg := &sync.WaitGroup{}
	Init(ctx, td, statikFS, wg)
	MapConf.MapName = "Test123"
	if err := SaveMapConf(); err != nil {
		t.Fatal(err)
	}
	defer cancel()
	backdb, err := getTmpDBFile()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(backdb)
	Backup.ConfigOnly = true
	err = backupDB(backdb)
	if err != nil {
		t.Fatal(err)
	}
	CloseDB()
	MapConf.MapName = ""
	err = openDB(backdb)
	if err != nil {
		t.Fatal(err)
	}
	if MapConf.MapName != "Test123" {
		t.Errorf("Backup MapName = '%s'", MapConf.MapName)
	}
	CloseDB()
}
