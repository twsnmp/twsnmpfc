package datastore

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"net/http"
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
	statikFS := http.Dir("../")
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

func TestMqttStatAutoCleanup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	statikFS := http.Dir("../")
	td, err := os.MkdirTemp("", "twsnmpfc_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)
	wg := &sync.WaitGroup{}
	Init(ctx, td, statikFS, wg)
	defer cancel()

	importTime := time.Now()

	// Add test MQTT stats
	UpdateMqttStat("client1", "192.168.1.1", "topic1", 100)
	UpdateMqttStat("client2", "192.168.1.2", "topic2", 200)

	// Artificially age the second stat
	k2 := getMqttStatKey("client2", "topic2")
	if v, ok := mqttStatMap.Load(k2); ok {
		if s, ok := v.(*MqttStatEnt); ok {
			s.Last = importTime.AddDate(0, 0, -3).UnixNano()
		}
	}
	SaveMqttStats()

	// Run cleanup with retention of 2 days
	deleted := DeleteOldMqttStats(2)
	if deleted != 1 {
		t.Errorf("expected 1 deleted MQTT stat, got %d", deleted)
	}

	// Verify that client1 remains and client2 is deleted
	k1 := getMqttStatKey("client1", "topic1")
	if _, ok := mqttStatMap.Load(k1); !ok {
		t.Error("client1 stat should not be deleted")
	}
	if _, ok := mqttStatMap.Load(k2); ok {
		t.Error("client2 stat should be deleted")
	}

	CloseDB()
}
