package datastore

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"go.etcd.io/bbolt"
)

type InfluxdbConfEnt struct {
	URL        string
	User       string
	Password   string
	DB         string
	Duration   string
	PollingLog string
	AIScore    string
}

func (ds *DataStore) SaveInfluxdbConfToDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(ds.InfluxdbConf)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("influxdbConf"), s)
	})
}

func (ds *DataStore) InitInfluxdb() error {
	if err := ds.dropInfluxdb(); err != nil {
		return err
	}
	return ds.setupInfluxdb()
}

func (ds *DataStore) setupInfluxdb() error {
	ds.closeInfluxdb()
	ds.muInfluxc.Lock()
	defer ds.muInfluxc.Unlock()
	if ds.InfluxdbConf.URL == "" {
		return nil
	}
	var err error
	conf := client.HTTPConfig{
		Addr:               ds.InfluxdbConf.URL,
		Timeout:            time.Second * 5,
		InsecureSkipVerify: true,
	}
	if ds.InfluxdbConf.User != "" && ds.InfluxdbConf.Password != "" {
		conf.Username = ds.InfluxdbConf.User
		conf.Password = ds.InfluxdbConf.Password
	}
	ds.influxc, err = client.NewHTTPClient(conf)
	if err != nil {
		ds.influxc = nil
		return err
	}
	return ds.checkInfluxdb()
}

func (ds *DataStore) checkInfluxdb() error {
	q := client.NewQuery("SHOW DATABASES", "", "")
	if response, err := ds.influxc.Query(q); err == nil && response.Error() == nil {
		for _, r := range response.Results {
			for _, s := range r.Series {
				for _, ns := range s.Values {
					for _, n := range ns {
						if name, ok := n.(string); ok {
							if name == ds.InfluxdbConf.DB {
								return nil
							}
						}
					}
				}
			}
		}
	} else {
		return err
	}
	qs := fmt.Sprintf(`CREATE DATABASE "%s"`, ds.InfluxdbConf.DB)
	if ds.InfluxdbConf.Duration != "" {
		qs += " WITH DURATION " + ds.InfluxdbConf.Duration
	}
	q = client.NewQuery(qs, "", "")
	if response, err := ds.influxc.Query(q); err != nil || response.Error() != nil {
		return err
	}
	return nil
}

func (ds *DataStore) dropInfluxdb() error {
	ds.muInfluxc.Lock()
	defer ds.muInfluxc.Unlock()
	if ds.influxc == nil {
		return nil
	}
	qs := fmt.Sprintf(`DROP DATABASE "%s"`, ds.InfluxdbConf.DB)
	q := client.NewQuery(qs, "", "")
	if response, err := ds.influxc.Query(q); err != nil || response.Error() != nil {
		return err
	}
	return nil
}

func (ds *DataStore) SendPollingLogToInfluxdb(pe *PollingEnt) error {
	ds.muInfluxc.Lock()
	defer ds.muInfluxc.Unlock()
	if ds.influxc == nil {
		return nil
	}
	n := ds.GetNode(pe.NodeID)
	if n == nil {
		return ErrInvalidID
	}
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  ds.InfluxdbConf.DB,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	// Create a point and add to batch
	tags := map[string]string{
		"map":       ds.MapConf.MapName,
		"node":      n.Name,
		"nodeID":    n.ID,
		"pollingID": pe.ID,
	}
	fields := map[string]interface{}{
		"numVal": pe.LastVal,
	}
	lr := make(map[string]string)
	if err := json.Unmarshal([]byte(pe.LastResult), &lr); err == nil {
		for k, v := range lr {
			if fv, err := strconv.ParseFloat(v, 64); err == nil {
				fields[k] = fv
			} else {
				fields[k] = v
			}
		}
	}
	pt, err := client.NewPoint(pe.Name, tags, fields, time.Unix(0, pe.LastTime))
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := ds.influxc.Write(bp); err != nil {
		return err
	}
	return nil
}

func (ds *DataStore) SendAIScoreToInfluxdb(pe *PollingEnt, res *AIResult) error {
	ds.muInfluxc.Lock()
	defer ds.muInfluxc.Unlock()
	if ds.influxc == nil {
		return nil
	}
	n := ds.GetNode(pe.NodeID)
	if n == nil {
		return ErrInvalidID
	}
	qs := fmt.Sprintf(`DROP SERIES FROM "AIScore" WHERE "pollingID" = "%s" `, pe.ID)
	q := client.NewQuery(qs, ds.InfluxdbConf.DB, "")
	if response, err := ds.influxc.Query(q); err != nil {
		log.Printf("sendAIScoreToInfluxdb err=%v", err)
		return err
	} else if response == nil {
		log.Printf("sendAIScoreToInfluxdb err=%v resp=nil", err)
		return err
	} else if response.Error() != nil {
		log.Printf("sendAIScoreToInfluxdb err=%v respError=%v", err, response.Error())
		return err
	}
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  ds.InfluxdbConf.DB,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	// Create a point and add to batch
	tags := map[string]string{
		"map":       ds.MapConf.MapName,
		"node":      n.Name,
		"nodeID":    n.ID,
		"pollingID": pe.ID,
	}
	for _, score := range res.ScoreData {
		if len(score) < 2 {
			continue
		}
		fields := map[string]interface{}{
			"AIScore": score[1],
		}
		pt, err := client.NewPoint("AIScore", tags, fields, time.Unix(int64(score[0]), 0))
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}
	// Write the batch
	if err := ds.influxc.Write(bp); err != nil {
		return err
	}
	return nil

}

func (ds *DataStore) closeInfluxdb() {
	ds.muInfluxc.Lock()
	defer ds.muInfluxc.Unlock()
	if ds.influxc == nil {
		return
	}
	ds.influxc.Close()
	ds.influxc = nil
}
