package backend

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/twsnmp/golof/lof"

	"github.com/twsnmp/twsnmpfc/datastore"
)

var aiBusy = false

const yasumi = `date,name
2019-01-01,元日
2019-01-14,成人の日
2019-02-11,建国記念の日
2019-03-21,春分の日
2019-04-29,昭和の日
2019-04-30,休日
2019-05-01,天皇の即位の日
2019-05-02,休日
2019-05-03,憲法記念日
2019-05-04,みどりの日
2019-05-05,こどもの日
2019-05-06,休日
2019-07-15,海の日
2019-08-11,山の日
2019-08-12,休日
2019-09-16,敬老の日
2019-09-23,秋分の日
2019-10-14,体育の日
2019-11-03,文化の日
2019-11-23,勤労感謝の日
2019-12-30,冬季休業
2019-12-31,冬季休業
2020-01-01,元日
2020-01-02,冬季休業
2020-01-03,冬季休業
2020-01-13,成人の日
2020-02-11,建国記念の日
2020-02-23,天皇の即位の日
2020-02-24,休日
2020-03-20,春分の日
2020-04-29,昭和の日
2020-05-03,憲法記念日
2020-05-04,みどりの日
2020-05-05,こどもの日
2020-05-06,休日
2020-07-23,海の日
2020-07-24,スポーツの日
2020-08-10,山の日
2020-09-21,敬老の日
2020-09-22,秋分の日
2020-11-03,文化の日
2020-11-23,勤労感謝の日
2021-1-1,元日
2021-1-11,成人の日
2021-2-11,建国記念の日
2021-2-23,天皇誕生日
2021-3-20,春分の日
2021-4-29,昭和の日
2021-5-3,憲法記念日
2021-5-4,みどりの日
2021-5-5,こどもの日
2021-7-22,海の日
2021-7-23,スポーツの日
2021-8-8,山の日
2021-8-9,休日
2021-9-20,敬老の日
2021-9-23,秋分の日
2021-11-3,文化の日
2021-11-23,勤労感謝の日
2022-1-1,元日
2022-1-10,成人の日
2022-2-11,建国記念の日
2022-2-23,天皇誕生日
2022-3-21,春分の日（予想）
2022-4-29,昭和の日
2022-5-3,憲法記念日
2022-5-4,みどりの日
2022-5-5,こどもの日
2022-7-18,海の日
2022-8-11,山の日
2022-9-19,敬老の日
2022-9-23,秋分の日（予想）
2022-10-10,スポーツの日
2022-11-3,文化の日
2022-11-23,勤労感謝の日
`

func aiBackend(ctx context.Context) {
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

type aiReq struct {
	PollingID string
	TimeStamp []int64
	Data      [][]float64
}

func makeYasumiMap() {
	for _, l := range strings.Split(yasumi, "\n") {
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
		log.Printf("loadAIReesult  id=%s err=%v", id, err)
		if err = datastore.DeleteAIResult(id); err != nil {
			log.Printf("loadAIReesult  id=%s err=%v", id, err)
		}
		return true
	}
	nextAIReqTimeMap[id] = last.LastTime
	return last.LastTime < time.Now().Unix()-60*60
}

func doAI(pe *datastore.PollingEnt) {
	if !checkLastAIResultTime(pe.ID) {
		return
	}
	req := &aiReq{
		PollingID: pe.ID,
	}
	err := makeAIData(req)
	if err != nil || len(req.Data) < 10 {
		log.Printf("doAI Skip No data %s %s", pe.ID, pe.Name)
		return
	}
	nextAIReqTimeMap[pe.ID] = time.Now().Unix() + 60*60
	log.Printf("doAI Start %s %s %d", pe.ID, pe.Name, len(req.Data))
	go calcAIScore(req)
}

const entLen = 20

func makeAIData(req *aiReq) error {
	p := datastore.GetPolling(req.PollingID)
	if p == nil {
		return fmt.Errorf("no polling")
	}
	keys := []string{}
	for k, v := range p.Result {
		if _, ok := v.(float64); !ok {
			continue
		}
		keys = append(keys, k)
	}
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
			ts := time.Unix(ct, 0)
			ent[0] = float64(ts.Hour())
			if _, ok := yasumiMap[ts.Format("2006-01-02")]; ok {
				ent[1] = 0.0
			} else {
				ent[1] = float64(ts.Weekday())
			}
			if count == 0.0 {
				continue
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

func calcAIScore(req *aiReq) {
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

func calcLOF(req *aiReq) *datastore.AIResult {
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
