package backend

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/montanaflynn/stats"
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

type aiDataFrame struct {
	Time []int64
	Data map[string][]float64
}

func (df aiDataFrame) Len() int {
	return len(df.Time)
}

func (df aiDataFrame) ColumnNames() []string {
	cols := []string{}
	for k := range df.Data {
		cols = append(cols, k)
	}
	return cols
}

func (df aiDataFrame) ToCSV(w io.Writer) error {
	cols := []string{}
	row := "time"
	for k := range df.Data {
		cols = append(cols, k)
		row += "," + k
	}
	if _, err := w.Write([]byte(row + "\n")); err != nil {
		return err
	}
	for i, t := range df.Time {
		row = time.Unix(t, 0).Format(time.RFC3339)
		for _, col := range cols {
			row += ","
			if v, ok := df.Data[col]; ok && len(v) > i {
				row += fmt.Sprintf("%f", v[i])
			}
		}
		if _, err := w.Write([]byte(row + "\n")); err != nil {
			return err
		}
	}
	return nil
}

type AIReq struct {
	PollingID string
	Df        aiDataFrame
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
	req.Df = aiDataFrame{
		Time: []int64{},
		Data: make(map[string][]float64),
	}
	req.Df.Data["hour"] = []float64{}
	req.Df.Data["weekday"] = []float64{}
	for _, k := range keys {
		req.Df.Data[k] = []float64{}
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
			req.Df.Time = append(req.Df.Time, ts.Unix())
			req.Df.Data["hour"] = append(req.Df.Data["hour"], float64(ts.Hour())/23)
			wd := float64(ts.Weekday())
			if _, ok := yasumiMap[ts.Format("2006-01-02")]; ok {
				wd = 0.0
			}
			req.Df.Data["weekday"] = append(req.Df.Data["weekday"], wd/6)
			for _, k := range keys {
				avg := ent[k] / count
				req.Df.Data[k] = append(req.Df.Data[k], avg)
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
		for j := range req.Df.Data[k] {
			if maxVals[k] > 0.0 {
				req.Df.Data[k][j] /= maxVals[k]
			} else {
				req.Df.Data[k][j] = 0.0
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
		for k := range req.Df.Data {
			if _, ok := colMap[k]; !ok {
				delete(req.Df.Data, k)
			}
		}
	}
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
		data[i] = make([]float64, len(cols))
	}
	for i, col := range cols {
		if v, ok := req.Df.Data[col]; ok {
			for j, d := range v {
				data[j][i] = d
			}
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
	for i := range r {
		score := ((10 * (float64(r[i]) - mean) / sd) + 50)
		res.ScoreData = append(res.ScoreData, []float64{float64(req.Df.Time[i]), score})
	}
	res.PollingID = req.PollingID
	res.LastTime = req.Df.Time[len(req.Df.Time)-1]
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
	for i := range r {
		score := ((10 * (float64(r[i]) - mean) / sd) + 50)
		res.ScoreData = append(res.ScoreData, []float64{float64(req.Df.Time[i]), score})
	}
	res.PollingID = req.PollingID
	res.LastTime = req.Df.Time[len(req.Df.Time)-1]
	return &res
}
