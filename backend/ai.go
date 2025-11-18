package backend

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/tobgu/qframe"
	"github.com/twsnmp/golof/lof"

	go_iforest "github.com/codegaudi/go-iforest"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func aiBackend(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start ai")
	timer := time.NewTicker(time.Second * 60)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			log.Println("stop ai")
			return
		case <-timer.C:
			checkAI()
		}
	}
}

type AIReq struct {
	PollingID string
	Df        qframe.QFrame
}

func makeYasumiMap() {
	for _, l := range strings.Split(datastore.Yasumi, "\n") {
		y := strings.Split(l, ",")
		if len(y) == 2 {
			if _, err := time.Parse("2006-01-02", y[0]); err == nil {
				yasumiMap[y[0]] = true
			}
		}
	}
}

func checkAI() {
	st := time.Now().Unix()
	datastore.ForEachPollings(func(pe *datastore.PollingEnt) bool {
		if pe.LogMode == datastore.LogModeAI {
			doAI(pe)
		}
		return time.Now().Unix()-st < 50
	})
}

func DeleteAIResult(id string) error {
	err := datastore.DeleteAIResult(id)
	if err == nil {
		nextAIReqTimeMap.Delete(id)
	}
	return err
}

var nextAIReqTimeMap sync.Map

func checkLastAIResultTime(id string) bool {
	if v, ok := nextAIReqTimeMap.Load(id); ok {
		if lt, ok := v.(int64); ok {
			return lt < time.Now().Unix()-60*60
		}
	}
	last, err := datastore.GetAIReesult(id)
	if err != nil {
		return true
	}
	nextAIReqTimeMap.Store(id, last.LastTime)
	return last.LastTime < time.Now().Unix()-60*60
}

func doAI(pe *datastore.PollingEnt) {
	if !checkLastAIResultTime(pe.ID) {
		return
	}
	req := &AIReq{
		PollingID: pe.ID,
	}
	err := MakeAIData(req)
	if err != nil {
		log.Printf("make ai data id=%s name=%s err=%v", pe.ID, pe.Name, err)
		return
	}
	if req.Df.Len() < 10 {
		log.Println("make ai data length < 10 skip")
		return
	}
	st := time.Now()
	calcAIScore(req, pe.AIMode)
	nextAIReqTimeMap.Store(pe.ID, time.Now().Unix())
	log.Printf("calc ai score id=%s name=%s len=%d dur=%v", pe.ID, pe.Name, req.Df.Len(), time.Since(st))
}

func getAIDataKeys(p *datastore.PollingEnt) []string {
	keys := []string{}
	if p.Type == "syslog" && p.Mode == "pri" {
		for i := 0; i < 256; i++ {
			keys = append(keys, fmt.Sprintf("pri_%d", i))
		}
		return keys
	}
	for k, v := range p.Result {
		// lastTimeは、測定データに含めない
		if k == "lastTime" {
			continue
		}
		if _, ok := v.(float64); !ok {
			continue
		}
		keys = append(keys, k)
	}
	return keys
}

func MakeAIData(req *AIReq) error {
	p := datastore.GetPolling(req.PollingID)
	if p == nil {
		return fmt.Errorf("no polling")
	}
	keys := getAIDataKeys(p)
	if len(keys) < 1 {
		return fmt.Errorf("no keys")
	}
	keys = append(keys, "state")
	data := make(map[string]interface{})
	data["time"] = []float64{}
	data["time_str"] = []string{}
	data["hour"] = []float64{}
	data["weekday"] = []float64{}
	for _, k := range keys {
		data[k] = []float64{}
	}
	logs := datastore.GetAllPollingLog(req.PollingID)
	if len(logs) < 1 {
		return fmt.Errorf("no logs")
	}
	st := 3600 * (time.Unix(0, logs[0].Time).Unix() / 3600)
	ent := make(map[string]float64)
	maxVals := make(map[string]float64)
	for _, k := range keys {
		ent[k] = 0.0
		maxVals[k] = 0.0
	}
	var count float64
	for _, l := range logs {
		ct := 3600 * (time.Unix(0, l.Time).Unix() / 3600)
		if st != ct {
			if count == 0.0 {
				// Dataがない場合はスキップする
				st = ct
				continue
			}
			ts := time.Unix(ct, 0)
			data["time"] = append(data["time"].([]float64), float64(ts.Unix()))
			data["time_str"] = append(data["time_str"].([]string), ts.Format(time.RFC3339))
			data["hour"] = append(data["hour"].([]float64), float64(ts.Hour())/23)
			wd := float64(ts.Weekday())
			if _, ok := yasumiMap[ts.Format("2006-01-02")]; ok {
				wd = 0.0
			}
			data["weekday"] = append(data["weekday"].([]float64), wd/6)
			for _, k := range keys {
				avg := ent[k] / count
				data[k] = append(data[k].([]float64), avg)
				if maxVals[k] < avg {
					maxVals[k] = avg
				}
			}
			for _, k := range keys {
				ent[k] = 0.0
			}
			st = ct
			count = 0.0
		}
		count += 1.0
		for _, k := range keys {
			if k == "state" {
				ent["state"] += getStateNum(l.State)
				continue
			}
			if v, ok := l.Result[k]; ok {
				if fv, ok := v.(float64); ok {
					ent[k] += fv
				}
			}
		}
	}
	for _, k := range keys {
		for j := range data[k].([]float64) {
			if maxVals[k] > 0.0 {
				data[k].([]float64)[j] /= maxVals[k]
			} else {
				data[k].([]float64)[j] = 0.0
			}
		}
	}
	if p.VectorCols != "" {
		colMap := make(map[string]bool)
		for _, c := range strings.Split(p.VectorCols, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				colMap[c] = true
			}
		}
		colMap["time"] = true
		colMap["time_str"] = true
		for k := range data {
			if _, ok := colMap[k]; !ok {
				delete(data, k)
			}
		}
	}
	req.Df = qframe.New(data)
	return nil
}

func getStateNum(s string) float64 {
	if s == "repair" || s == "normal" {
		return 1.0
	}
	if s == "unknown" {
		return 0.5
	}
	return 0.0
}

func calcAIScore(req *AIReq, aiMode string) {
	var res *datastore.AIResult
	if aiMode == "" {
		aiMode = datastore.MapConf.AIMode
	}
	switch aiMode {
	case "lof":
		res = calcLOF(req)
	default:
		res = calcIForest(req)
	}
	if len(res.ScoreData) < 1 {
		return
	}
	if err := datastore.SaveAIResult(res); err != nil {
		log.Printf("save ai result err=%v", err)
		return
	}
	pe := datastore.GetPolling(req.PollingID)
	if pe == nil {
		return
	}
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		return
	}
	if len(res.ScoreData) > 0 {
		ls := res.ScoreData[len(res.ScoreData)-1][1]
		if ls > float64(datastore.MapConf.AIThreshold) {
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:     "ai",
				Level:    datastore.MapConf.AILevel,
				NodeID:   pe.NodeID,
				NodeName: n.Name,
				Event:    fmt.Sprintf("AI分析レポート:%s(%s):%f", pe.Name, pe.Type, ls),
			})
		}
		if datastore.InfluxdbConf.AIScore == "send" {
			if err := datastore.SendAIScoreToInfluxdb(pe, res); err != nil {
				log.Printf("send ai score to influxdb id=%s name=%s err=%v", pe.ID, pe.Name, err)
			}
		}
	}
}

func getSampleData(req *AIReq) [][]float64 {
	cols := req.Df.ColumnNames()
	data := make([][]float64, req.Df.Len())
	for i := range data {
		data[i] = make([]float64, len(cols)-1)
	}
	i := 0
	for _, col := range cols {
		if col == "time" || col == "time_str" {
			continue
		}
		if v, err := req.Df.FloatView(col); err == nil {
			for j, d := range v.Slice() {
				data[j][i] = d
			}
			i++
		}
	}
	return data
}

func calcLOF(req *AIReq) *datastore.AIResult {
	res := datastore.AIResult{}
	data := getSampleData(req)
	samples := lof.GetSamplesFromFloat64s(data)
	lofGetter := lof.NewLOF(5)
	if err := lofGetter.Train(samples); err != nil {
		log.Printf("calc lof id=%s err=%v", req.PollingID, err)
		return &res
	}
	r := make([]float64, len(samples))

	for i, s := range samples {
		r[i] = lofGetter.GetLOF(s, "fast")
	}
	max, err := stats.Max(r)
	if err != nil {
		return &res
	}
	min, err := stats.Min(r)
	if err != nil {
		return &res
	}
	diff := max - min
	if diff == 0 {
		return &res
	}
	for i := range r {
		r[i] /= diff
		r[i] *= 100.0
	}
	mean, err := stats.Mean(r)
	if err != nil {
		return &res
	}
	sd, err := stats.StandardDeviation(r)
	if err != nil {
		return &res
	}
	times := []int64{}
	if v, err := req.Df.FloatView("time"); err != nil {
		log.Println(err)
		return &res
	} else {
		for _, t := range v.Slice() {
			times = append(times, int64(t))
		}
	}
	for i := range r {
		score := ((10 * (float64(r[i]) - mean) / sd) + 50)
		res.ScoreData = append(res.ScoreData, []float64{float64(times[i]), score})
	}
	res.PollingID = req.PollingID
	res.LastTime = times[len(times)-1]
	return &res
}

func calcIForest(req *AIReq) *datastore.AIResult {
	res := datastore.AIResult{}
	sub := 256
	if req.Df.Len() < sub {
		sub = req.Df.Len() / 2
		log.Printf("IForest subSample=%d", sub)
	}
	data := getSampleData(req)
	iforest, err := go_iforest.NewIForest(data, 1000, sub)
	if err != nil {
		log.Printf("NewIForest err=%v", err)
		return &res
	}
	r := make([]float64, len(data))
	for i, v := range data {
		r[i] = iforest.CalculateAnomalyScore(v)
	}
	max, err := stats.Max(r)
	if err != nil {
		return &res
	}
	min, err := stats.Min(r)
	if err != nil {
		return &res
	}
	diff := max - min
	if diff == 0 {
		return &res
	}
	for i := range r {
		r[i] /= diff
		r[i] *= 100.0
	}
	mean, err := stats.Mean(r)
	if err != nil {
		return &res
	}
	sd, err := stats.StandardDeviation(r)
	if err != nil {
		return &res
	}
	times := []int64{}
	if v, err := req.Df.FloatView("time"); err != nil {
		log.Println(err)
		return &res
	} else {
		for _, t := range v.Slice() {
			times = append(times, int64(t))
		}
	}
	for i := range r {
		score := ((10 * (float64(r[i]) - mean) / sd) + 50)
		res.ScoreData = append(res.ScoreData, []float64{float64(times[i]), score})
	}
	res.PollingID = req.PollingID
	res.LastTime = times[len(times)-1]
	return &res
}
