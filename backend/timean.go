package backend

import (
	"fmt"
	"math/cmplx"
	"time"

	"github.com/chewxy/stl"
	"github.com/mjibson/go-dsp/fft"

	"github.com/twsnmp/twsnmpfc/datastore"
)

type TimeAnalyzedPollingLog struct {
	PX2        int64
	TimeH      []int64
	TimePX2    []int64
	DataMapH   map[string][]float64
	StlMapH    map[string]stl.Result
	DataMapPX2 map[string][]float64
	StlMapPX2  map[string]stl.Result
	FFTH       map[string][][]float64
	FFTPX2     map[string][][]float64
}

func TimeAnalyzePollingLog(id string) (TimeAnalyzedPollingLog, error) {
	r := TimeAnalyzedPollingLog{
		DataMapH:   make(map[string][]float64),
		StlMapH:    make(map[string]stl.Result),
		DataMapPX2: make(map[string][]float64),
		StlMapPX2:  make(map[string]stl.Result),
		FFTH:       make(map[string][][]float64),
		FFTPX2:     make(map[string][][]float64),
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
	if len(logs) < 10 {
		return r, fmt.Errorf("not enough logs %d", len(logs))
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
			r.TimeH = append(r.TimeH, time.Unix(sth, 0).Unix())
			for _, k := range keys {
				r.DataMapH[k] = append(r.DataMapH[k], entH[k])
			}
			sth += 3600
			for ; sth < cth; sth += 3600 {
				r.TimeH = append(r.TimeH, time.Unix(sth, 0).Unix())
				for _, k := range keys {
					r.DataMapH[k] = append(r.DataMapH[k], entH[k])
				}
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
			r.TimePX2 = append(r.TimePX2, time.Unix(stpx2, 0).Unix())
			for _, k := range keys {
				r.DataMapPX2[k] = append(r.DataMapPX2[k], entPX2[k])
			}
			stpx2 += r.PX2
			for ; stpx2 < ctpx2; stpx2 += r.PX2 {
				r.TimePX2 = append(r.TimePX2, time.Unix(stpx2, 0).Unix())
				for _, k := range keys {
					r.DataMapPX2[k] = append(r.DataMapPX2[k], 0.0)
				}
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
	dpx2 := int(3600 / r.PX2)
	for _, k := range keys {
		if len(r.DataMapH[k]) >= 24 {
			r.StlMapH[k] = stl.Decompose(r.DataMapH[k], 24, 24*3-1, stl.Additive(), stl.WithRobustIter(2), stl.WithIter(2))
		}
		if len(r.DataMapPX2[k]) >= dpx2 {
			r.StlMapPX2[k] = stl.Decompose(r.DataMapPX2[k], dpx2, dpx2*3-1, stl.Additive(), stl.WithRobustIter(2), stl.WithIter(2))
		}
		r.FFTH[k] = getFFTData(float64(1/3600.0), r.DataMapH[k])
		r.FFTPX2[k] = getFFTData(1.0/float64(r.PX2), r.DataMapPX2[k])
	}
	return r, nil
}

func getFFTData(sampleRate float64, data []float64) [][]float64 {
	ret := [][]float64{}
	// 高速フーリエ変換する
	fftdata := fft.FFTReal(data)
	// 両方向のスペクトラムを計算する
	var spectrum2 []float64
	length := float64(len(data))
	for i := range fftdata {
		// FFTの結果を複素数から実数に変換する
		spectrum2 = append(spectrum2, cmplx.Abs(fftdata[i])/length)
	}
	// 片側のスペクトラムだけにする
	spectrum1 := spectrum2[0 : len(spectrum2)/2]
	for i := range spectrum1 {
		spectrum1[i] = spectrum1[i] * 2
	}
	// 周波数とスペクトラムの関係のデータを作成する
	spectrumLength := float64(len(spectrum2))
	for i, v := range spectrum1 {
		ret = append(ret, []float64{float64(i) * sampleRate / spectrumLength, v})
	}
	return ret
}
