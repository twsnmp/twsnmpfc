package backend

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chewxy/stl"
	"github.com/montanaflynn/stats"
	"github.com/twsnmp/golof/lof"

	"github.com/twsnmp/twsnmpfc/datastore"
)

func aiBackend(ctx context.Context) {
	log.Println("start ai backend")
	timer := time.NewTicker(time.Second * 60)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			checkAI()
		}
	}
}

type AIReq struct {
	PollingID string
	TimeStamp []int64
	Data      [][]float64
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
	datastore.ForEachPollings(func(pe *datastore.PollingEnt) bool {
		if pe.LogMode == datastore.LogModeAI {
			doAI(pe)
		}
		return true
	})
}

func DeleteAIResult(id string) error {
	err := datastore.DeleteAIResult(id)
	if err == nil {
		delete(nextAIReqTimeMap, id)
	}
	return err
}

var nextAIReqTimeMap = make(map[string]int64)

func checkLastAIResultTime(id string) bool {
	if lt, ok := nextAIReqTimeMap[id]; ok {
		return lt < time.Now().Unix()-60*60
	}
	last, err := datastore.GetAIReesult(id)
	if err != nil {
		return true
	}
	nextAIReqTimeMap[id] = last.LastTime
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
		log.Printf("doAI skip MakeAIData error %s %s err=%v", pe.ID, pe.Name, err)
		return
	}
	if err != nil || len(req.Data) < 10 {
		log.Printf("doAI skip AIData length  %s %s len=%d", pe.ID, pe.Name, len(req.Data))
		return
	}
	nextAIReqTimeMap[pe.ID] = time.Now().Unix() + 60*60
	log.Printf("doAI start %s %s %d", pe.ID, pe.Name, len(req.Data))
	go calcAIScore(req)
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
			log.Printf("skip key %s=%v", k, v)
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
	logs := datastore.GetAllPollingLog(req.PollingID)
	if len(logs) < 1 {
		return fmt.Errorf("no logs")
	}
	entLen := len(keys) + 3
	st := 3600 * (time.Unix(0, logs[0].Time).Unix() / 3600)
	ent := make([]float64, entLen)
	maxVals := make([]float64, entLen)
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
			ent[0] = float64(ts.Hour())
			if _, ok := yasumiMap[ts.Format("2006-01-02")]; ok {
				ent[1] = 0.0
			} else {
				ent[1] = float64(ts.Weekday())
			}
			for i := 0; i < len(ent); i++ {
				if i >= 3 {
					ent[i] /= count
				}
				if maxVals[i] < ent[i] {
					maxVals[i] = ent[i]
				}
			}
			req.TimeStamp = append(req.TimeStamp, ts.Unix())
			req.Data = append(req.Data, ent)
			ent = make([]float64, entLen)
			st = ct
			count = 0.0
		}
		count += 1.0
		ent[3] += getStateNum(l.State)
		for i, k := range keys {
			if v, ok := l.Result[k]; ok {
				if fv, ok := v.(float64); ok {
					ent[i+3] += fv
				}
			}
		}
	}
	for i := range req.Data {
		for j := range req.Data[i] {
			if maxVals[j] > 0.0 {
				req.Data[i][j] /= maxVals[j]
			} else {
				req.Data[i][j] = 0.0
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

func calcAIScore(req *AIReq) {
	res := calcLOF(req)
	if len(res.ScoreData) < 1 {
		return
	}
	if err := datastore.SaveAIResult(res); err != nil {
		log.Printf("saveAIResult err=%v", err)
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
				log.Printf("sendAIScoreToInfluxdb err=%v", err)
			}
		}
	}
}

func calcLOF(req *AIReq) *datastore.AIResult {
	res := datastore.AIResult{}
	samples := lof.GetSamplesFromFloat64s(req.Data)
	lofGetter := lof.NewLOF(5)
	if err := lofGetter.Train(samples); err != nil {
		log.Printf("calcLOF err=%v", err)
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
		// r[i] = (1.0 - r[i]) * 100.0
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
		res.ScoreData = append(res.ScoreData, []float64{float64(req.TimeStamp[i]), score})
	}
	res.PollingID = req.PollingID
	res.LastTime = req.TimeStamp[len(req.TimeStamp)-1]
	return &res
}

type TimeAnalyzedPollingLog struct {
	PX2        int64
	TimeH      []int64
	TimePX2    []int64
	DataMapH   map[string][]float64
	StlMapH    map[string]stl.Result
	DataMapPX2 map[string][]float64
	StlMapPX2  map[string]stl.Result
}

func TimeAnalyzePollingLog(id string) (TimeAnalyzedPollingLog, error) {
	r := TimeAnalyzedPollingLog{
		DataMapH:   make(map[string][]float64),
		StlMapH:    make(map[string]stl.Result),
		DataMapPX2: make(map[string][]float64),
		StlMapPX2:  make(map[string]stl.Result),
	}
	p := datastore.GetPolling(id)
	if p == nil {
		return r, fmt.Errorf("no polling")
	}
	keys := getAIDataKeys(p)
	if len(keys) < 1 {
		return r, fmt.Errorf("no keys")
	}
	logs := datastore.GetAllPollingLog(id)
	if len(logs) < 1 {
		return r, fmt.Errorf("no logs")
	}
	r.PX2 = int64(p.PollInt * 2)
	entH := make(map[string]float64)
	entPX2 := make(map[string]float64)
	var countH float64
	var countPX2 float64
	sth := 3600 * (time.Unix(0, logs[0].Time).Unix() / 3600)
	stpx2 := r.PX2 * (time.Unix(0, logs[0].Time).Unix() / r.PX2)
	for _, k := range keys {
		r.DataMapH[k] = []float64{}
		r.DataMapPX2[k] = []float64{}
		entH[k] = 0.0
		entPX2[k] = 0.0
	}
	for _, l := range logs {
		cth := 3600 * (time.Unix(0, l.Time).Unix() / 3600)
		if sth != cth {
			if countH > 0.0 {
				for _, k := range keys {
					entH[k] /= countH
				}
			}
			for ; sth < cth-3600; sth += 3600 {
				r.TimeH = append(r.TimeH, time.Unix(sth, 0).Unix())
				for _, k := range keys {
					r.DataMapH[k] = append(r.DataMapH[k], entH[k])
				}
			}
			r.TimeH = append(r.TimeH, time.Unix(cth, 0).Unix())
			for _, k := range keys {
				r.DataMapH[k] = append(r.DataMapH[k], entH[k])
			}
			for _, k := range keys {
				entH[k] = 0.0
			}
			sth = cth
			countH = 0.0
		}
		ctpx2 := r.PX2 * (time.Unix(0, l.Time).Unix() / r.PX2)
		if stpx2 != ctpx2 {
			if countPX2 > 0.0 {
				for _, k := range keys {
					entPX2[k] /= countPX2
				}
			}
			for ; stpx2 < ctpx2-r.PX2; stpx2 += r.PX2 {
				r.TimePX2 = append(r.TimePX2, time.Unix(stpx2, 0).Unix())
				for _, k := range keys {
					r.DataMapPX2[k] = append(r.DataMapPX2[k], 0.0)
				}
			}
			r.TimePX2 = append(r.TimePX2, time.Unix(ctpx2, 0).Unix())
			for _, k := range keys {
				r.DataMapPX2[k] = append(r.DataMapPX2[k], entPX2[k])
			}
			for _, k := range keys {
				entPX2[k] = 0.0
			}
			stpx2 = ctpx2
			countPX2 = 0.0
		}
		countH += 1.0
		countPX2 += 1.0
		for _, k := range keys {
			if v, ok := l.Result[k]; ok {
				if fv, ok := v.(float64); ok {
					entH[k] += fv
					entPX2[k] += fv
				}
			}
		}
	}
	dpx2 := int((24 * 3600) / r.PX2)
	for _, k := range keys {
		r.StlMapH[k] = stl.Decompose(r.DataMapH[k], 24, 24*3-1, stl.Additive(), stl.WithRobustIter(2), stl.WithIter(2))
		r.StlMapPX2[k] = stl.Decompose(r.DataMapPX2[k], dpx2, dpx2*3-1, stl.Additive(), stl.WithRobustIter(2), stl.WithIter(2))
	}
	return r, nil
}
