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

func TestMigrateMapSize(t *testing.T) {
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

	// 1. はみ出すオブジェクトがない場合のテスト (MapSize = 2)
	// X: 3000 (最大3000), Y: 4000 (最大4000) => 共に5000以内なので縮小されないはず
	MapConf.MapSize = 2
	SaveMapConf()

	n1 := &NodeEnt{Name: "Node1", X: 1000, Y: 4000}
	if err := AddNode(n1); err != nil {
		t.Fatal(err)
	}
	di1 := &DrawItemEnt{Type: DrawItemTypeRect, X: 2000, Y: 3000, W: 100, H: 200, Size: 16}
	if err := AddDrawItem(di1); err != nil {
		t.Fatal(err)
	}
	net1 := &NetworkEnt{IP: "192.168.1.0", X: 3000, Y: 2000}
	if err := AddNetwork(net1); err != nil {
		t.Fatal(err)
	}

	migrateMapSize()

	if MapConf.MapSize != 1 {
		t.Errorf("expected MapSize 1, got %d", MapConf.MapSize)
	}

	nCheck1 := GetNode(n1.ID)
	if nCheck1.X != 1000 || nCheck1.Y != 4000 {
		t.Errorf("Node1 should not be scaled. expected (1000, 4000), got (%d, %d)", nCheck1.X, nCheck1.Y)
	}

	diCheck1 := GetDrawItem(di1.ID)
	if diCheck1.X != 2000 || diCheck1.Y != 3000 {
		t.Errorf("DrawItem1 should not be scaled. expected (2000, 3000), got (%d, %d)", diCheck1.X, diCheck1.Y)
	}

	netCheck1 := GetNetwork(net1.ID)
	if netCheck1.X != 3000 || netCheck1.Y != 2000 {
		t.Errorf("Network1 should not be scaled. expected (3000, 2000), got (%d, %d)", netCheck1.X, netCheck1.Y)
	}

	// 2. はみ出すオブジェクトがある場合のテスト (MapSize = 3)
	// 最大X: 8000, 最大Y: 6000 => 5000を超えるため縮小されるはず
	// scaleX = 4968.0 / 8000.0 = 0.621
	// scaleY = 4968.0 / 6000.0 = 0.828
	MapConf.MapSize = 3
	SaveMapConf()

	n2 := &NodeEnt{Name: "Node2", X: 8000, Y: 3000}
	if err := AddNode(n2); err != nil {
		t.Fatal(err)
	}
	di2 := &DrawItemEnt{Type: DrawItemTypeRect, X: 4000, Y: 5800, W: 100, H: 200, Size: 24} // Y+H = 6000
	if err := AddDrawItem(di2); err != nil {
		t.Fatal(err)
	}
	net2 := &NetworkEnt{IP: "192.168.2.0", X: 3000, Y: 2000}
	if err := AddNetwork(net2); err != nil {
		t.Fatal(err)
	}

	migrateMapSize()

	if MapConf.MapSize != 1 {
		t.Errorf("expected MapSize 1, got %d", MapConf.MapSize)
	}

	nCheck2 := GetNode(n2.ID)
	// 8000 * 0.621 = 4968
	// 3000 * 0.828 = 2484
	if nCheck2.X != 4968 || nCheck2.Y != 2484 {
		t.Errorf("Node2 should be scaled to (4968, 2484), got (%d, %d)", nCheck2.X, nCheck2.Y)
	}

	diCheck2 := GetDrawItem(di2.ID)
	// X: 4000 * 0.621 = 2484, Y: 5800 * 0.828 = 4802
	// W: 100 * 0.621 = 62, H: 200 * 0.828 = 165
	// Size: 24 * 0.621 = 14
	if diCheck2.X != 2484 || diCheck2.Y != 4802 || diCheck2.W != 62 || diCheck2.H != 165 || diCheck2.Size != 14 {
		t.Errorf("DrawItem2 should be scaled. expected X:2484 Y:4802 W:62 H:165 Size:14, got X:%d Y:%d W:%d H:%d Size:%d",
			diCheck2.X, diCheck2.Y, diCheck2.W, diCheck2.H, diCheck2.Size)
	}

	netCheck2 := GetNetwork(net2.ID)
	// X: 3000 * 0.621 = 1863, Y: 2000 * 0.828 = 1656
	if netCheck2.X != 1863 || netCheck2.Y != 1656 {
		t.Errorf("Network2 should be scaled to (1863, 1656), got (%d, %d)", netCheck2.X, netCheck2.Y)
	}

	CloseDB()
}


