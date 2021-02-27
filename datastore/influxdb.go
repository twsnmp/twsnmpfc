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

func SaveInfluxdbConfToDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(InfluxdbConf)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("influxdbConf"), s)
	})
}

func InitInfluxdb() error {
	if err := dropInfluxdb(); err != nil {
		return err
	}
	return setupInfluxdb()
}

func setupInfluxdb() error {
	closeInfluxdb()
	muInfluxc.Lock()
	defer muInfluxc.Unlock()
	if InfluxdbConf.URL == "" {
		return nil
	}
	var err error
	conf := client.HTTPConfig{
		Addr:               InfluxdbConf.URL,
		Timeout:            time.Second * 5,
		InsecureSkipVerify: true,
	}
	if InfluxdbConf.User != "" && InfluxdbConf.Password != "" {
		conf.Username = InfluxdbConf.User
		conf.Password = InfluxdbConf.Password
	}
	influxc, err = client.NewHTTPClient(conf)
	if err != nil {
		influxc = nil
		return err
	}
	return checkInfluxdb()
}

func checkInfluxdb() error {
	q := client.NewQuery("SHOW DATABASES", "", "")
	if response, err := influxc.Query(q); err == nil && response.Error() == nil {
		for _, r := range response.Results {
			for _, s := range r.Series {
				for _, ns := range s.Values {
					for _, n := range ns {
						if name, ok := n.(string); ok {
							if name == InfluxdbConf.DB {
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
	qs := fmt.Sprintf(`CREATE DATABASE "%s"`, InfluxdbConf.DB)
	if InfluxdbConf.Duration != "" {
		qs += " WITH DURATION " + InfluxdbConf.Duration
	}
	q = client.NewQuery(qs, "", "")
	if response, err := influxc.Query(q); err != nil || response.Error() != nil {
		return err
	}
	return nil
}

func dropInfluxdb() error {
	muInfluxc.Lock()
	defer muInfluxc.Unlock()
	if influxc == nil {
		return nil
	}
	qs := fmt.Sprintf(`DROP DATABASE "%s"`, InfluxdbConf.DB)
	q := client.NewQuery(qs, "", "")
	if response, err := influxc.Query(q); err != nil || response.Error() != nil {
		return err
	}
	return nil
}

func SendPollingLogToInfluxdb(pe *PollingEnt) error {
	muInfluxc.Lock()
	defer muInfluxc.Unlock()
	if influxc == nil {
		return nil
	}
	n := GetNode(pe.NodeID)
	if n == nil {
		return ErrInvalidID
	}
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  InfluxdbConf.DB,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	// Create a point and add to batch
	tags := map[string]string{
		"map":       MapConf.MapName,
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
	if err := influxc.Write(bp); err != nil {
		return err
	}
	return nil
}

func SendAIScoreToInfluxdb(pe *PollingEnt, res *AIResult) error {
	muInfluxc.Lock()
	defer muInfluxc.Unlock()
	if influxc == nil {
		return nil
	}
	n := GetNode(pe.NodeID)
	if n == nil {
		return ErrInvalidID
	}
	qs := fmt.Sprintf(`DROP SERIES FROM "AIScore" WHERE "pollingID" = "%s" `, pe.ID)
	q := client.NewQuery(qs, InfluxdbConf.DB, "")
	if response, err := influxc.Query(q); err != nil {
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
		Database:  InfluxdbConf.DB,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	// Create a point and add to batch
	tags := map[string]string{
		"map":       MapConf.MapName,
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
	if err := influxc.Write(bp); err != nil {
		return err
	}
	return nil

}

func closeInfluxdb() {
	muInfluxc.Lock()
	defer muInfluxc.Unlock()
	if influxc == nil {
		return
	}
	influxc.Close()
	influxc = nil
}
