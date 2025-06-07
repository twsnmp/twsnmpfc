package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"

	"github.com/twsnmp/twsnmpfc/datastore"
)

var OTelGRPCPort = 4717
var OTelHTTPPort = 4718
var OTelCert = ""
var OTelKey = ""
var OTelCA = ""
var traceMap sync.Map
var otelFromMap sync.Map
var limitFrom bool

func oteld(stopCh chan bool) {
	log.Printf("start oteld")
	datastore.LoadOTelMetric()
	setOTelFrom()
	f := otlpreceiver.NewFactory()
	config := f.CreateDefaultConfig()
	if c, ok := config.(*otlpreceiver.Config); ok {
		c.HTTP.ServerConfig.Endpoint = fmt.Sprintf(":%d", OTelHTTPPort)
		c.GRPC.NetAddr.Endpoint = fmt.Sprintf(":%d", OTelGRPCPort)
		if OTelCert != "" && OTelKey != "" {
			c.HTTP.ServerConfig.TLSSetting = &configtls.ServerConfig{
				Config: configtls.Config{
					CertFile: OTelCert,
					KeyFile:  OTelKey,
				},
			}
			c.HTTP.ServerConfig.TLSSetting.KeyFile = OTelKey
			c.GRPC.TLSSetting = &configtls.ServerConfig{
				Config: configtls.Config{
					CertFile: OTelCert,
					KeyFile:  OTelKey,
				},
			}
			if OTelCA != "" {
				c.HTTP.ServerConfig.TLSSetting.ClientCAFile = OTelCA
				c.GRPC.TLSSetting.ClientCAFile = OTelCA
				log.Println("otlp mTLS")
			} else {
				log.Println("otlp TLS")
			}
		} else {
			log.Println("otlp not TLS")
		}
	}
	componentID := component.MustNewID("otlp")
	ctx := context.Background()
	mr, err := consumer.NewMetrics(handleMetrics)
	if err != nil {
		log.Printf("oteld err=%v", err)
	}
	_, err = f.CreateMetrics(ctx,
		receiver.Settings{
			ID:                componentID,
			TelemetrySettings: componenttest.NewNopTelemetrySettings(),
		},
		config, mr)
	if err != nil {
		log.Printf("Failed to create receiver otlp: %v", err)
	}
	tr, err := consumer.NewTraces(handleTraces)
	if err != nil {
		log.Printf("oteld err=%v", err)
	}
	_, err = f.CreateTraces(ctx,
		receiver.Settings{
			ID:                componentID,
			TelemetrySettings: componenttest.NewNopTelemetrySettings(),
		},
		config, tr)
	if err != nil {
		log.Printf("Failed to create receiver otlp: %v", err)
	}
	lr, err := consumer.NewLogs(handleLogs)
	if err != nil {
		log.Printf("oteld err=%v", err)
	}
	logReceiver, err := f.CreateLogs(ctx,
		receiver.Settings{
			ID:                componentID,
			TelemetrySettings: componenttest.NewNopTelemetrySettings(),
		},
		config, lr)
	if err != nil {
		log.Printf("Failed to create receiver otlp: %v", err)
	}
	err = logReceiver.Start(ctx, nil)
	if err != nil {
		log.Printf("oteld err=%v", err)
	}
	timer := time.NewTicker(time.Minute)
	lastSave := int64(0)
	for {
		select {
		case <-stopCh:
			if logReceiver != nil {
				logReceiver.Shutdown(ctx)
			}
			traceMap.Clear()
			datastore.SaveOTelMetric()
			log.Printf("stop oteld")
			return
		case <-timer.C:
			{
				setOTelFrom()
				delList := []string{}
				saveList := []*datastore.OTelTraceEnt{}
				et := time.Now().Add(time.Duration(-datastore.MapConf.OTelRetention-1) * time.Hour).UnixNano()
				maxLast := int64(0)
				traceMap.Range(func(key any, value any) bool {
					if t, ok := value.(*datastore.OTelTraceEnt); ok {
						if t.Last < et {
							delList = append(delList, key.(string))
						}
						if t.Last > lastSave {
							if maxLast < t.Last {
								maxLast = t.Last
							}
							saveList = append(saveList, t)
						}
					}
					return true
				})
				if len(saveList) > 0 {
					lastSave = maxLast
					datastore.UpdateOTelTrace(saveList)
				}
				if len(delList) > 0 {
					for _, tid := range delList {
						traceMap.Delete(tid)
					}
					log.Printf("delete otel trace len=%d ", len(delList))
				}
			}
		}
	}
}

var lastOTelFrom = ""

func setOTelFrom() {
	if datastore.MapConf.OTelFrom == lastOTelFrom {
		return
	}
	lastOTelFrom = datastore.MapConf.OTelFrom
	a := strings.Split(datastore.MapConf.OTelFrom, ",")
	otelFromMap.Clear()
	limitFrom = false
	for _, f := range a {
		if f != "" {
			otelFromMap.Store(f, true)
			limitFrom = true
		}
	}
}

func handleMetrics(ctx context.Context, md pmetric.Metrics) error {
	f := client.FromContext(ctx)
	service := "unknown"
	host := f.Addr.String()
	if limitFrom {
		if _, ok := otelFromMap.Load(host); !ok {
			return nil
		}
	}
	for _, rm := range md.ResourceMetrics().All() {
		if v, ok := rm.Resource().Attributes().Get("host.name"); ok {
			host = v.AsString()
		}
		if v, ok := rm.Resource().Attributes().Get("service.name"); ok {
			service = v.AsString()
		}
		for _, sm := range rm.ScopeMetrics().All() {
			for _, m := range sm.Metrics().All() {
				metric := datastore.FindOTelMetric(host, service, sm.Scope().Name(), m.Name())
				if metric != nil {
					metric.Count++
					metric.Last = time.Now().UnixNano()
				} else {
					metric = &datastore.OTelMetricEnt{
						Host:        host,
						Service:     service,
						Scope:       sm.Scope().Name(),
						Name:        m.Name(),
						Type:        m.Type().String(),
						First:       time.Now().UnixNano(),
						Last:        time.Now().UnixNano(),
						Description: m.Description(),
						Unit:        m.Unit(),
						Count:       1,
					}
					datastore.AddOTelMetric(metric)
				}
				addDataPoints(metric, &m)
			}
		}
	}
	return nil
}

func handleTraces(ctx context.Context, td ptrace.Traces) error {
	f := client.FromContext(ctx)
	host := f.Addr.String()
	if limitFrom {
		if _, ok := otelFromMap.Load(host); !ok {
			return nil
		}
	}
	service := "unknown"
	for _, rs := range td.ResourceSpans().All() {
		if v, ok := rs.Resource().Attributes().Get("service.name"); ok {
			service = v.AsString()
		}
		if v, ok := rs.Resource().Attributes().Get("host.name"); ok {
			host = v.AsString()
		}
		for _, ss := range rs.ScopeSpans().All() {
			scope := ss.Scope().Name()
			for _, s := range ss.Spans().All() {
				tid := s.TraceID().String()
				var trace *datastore.OTelTraceEnt
				if v, ok := traceMap.Load(tid); ok {
					if p, ok := v.(*datastore.OTelTraceEnt); ok {
						trace = p
					}
				}
				if trace == nil {
					trace = &datastore.OTelTraceEnt{
						TraceID: tid,
						Bucket:  time.Now().Format("2006-01-02T15:04"),
						Spans:   []datastore.OTelTraceSpanEnt{},
					}
					traceMap.Store(tid, trace)
				}
				st := s.StartTimestamp().AsTime().UnixNano()
				et := s.EndTimestamp().AsTime().UnixNano()
				dur := float64(et-st) / (1000.0 * 1000.0 * 1000.0)
				trace.Spans = append(trace.Spans, datastore.OTelTraceSpanEnt{
					Name:         s.Name(),
					Service:      service,
					Host:         host,
					Scope:        scope,
					Attributes:   getAttributes(s.Attributes().AsRaw()),
					SpanID:       s.SpanID().String(),
					ParentSpanID: s.ParentSpanID().String(),
					Start:        st,
					End:          et,
					Dur:          dur,
				})
				if trace.Start == 0 || trace.Start > st {
					trace.Start = st
				}
				if trace.End < et {
					trace.End = et
				}
				if trace.Dur < dur {
					trace.Dur = dur
				}
				trace.Last = time.Now().UnixNano()
			}
		}
	}
	return nil
}

func handleLogs(ctx context.Context, ld plog.Logs) error {
	f := client.FromContext(ctx)
	host := f.Addr.String()
	if limitFrom {
		if _, ok := otelFromMap.Load(host); !ok {
			return nil
		}
	}
	service := "unknown"
	for _, rl := range ld.ResourceLogs().All() {
		if v, ok := rl.Resource().Attributes().Get("host.name"); ok {
			host = v.AsString()
		}
		if v, ok := rl.Resource().Attributes().Get("service.name"); ok {
			service = v.AsString()
		}
		for _, sl := range rl.ScopeLogs().All() {
			scope := sl.Scope().Name()

			for _, l := range sl.LogRecords().All() {
				logMap := make(map[string]any)
				logMap["content"] = fmt.Sprintf("service=%s scope=%s tarceid=%s spanid=%s event=%s severity=%s %s",
					service, scope, l.TraceID().String(), l.SpanID().String(), l.EventName(), l.SeverityText(), l.Body().AsString())
				logMap["tag"] = "otel"
				logMap["severity"] = getOTelLogSeverity(int(l.SeverityNumber()))
				logMap["facility"] = float64(16)
				logMap["hostname"] = host
				if j, err := json.Marshal(&logMap); err == nil {
					logCh <- &datastore.LogEnt{
						Time: l.Timestamp().AsTime().UnixNano(),
						Type: "syslog",
						Log:  string(j),
					}
				}
			}
		}
	}
	return nil
}

func getOTelLogSeverity(s int) float64 {
	switch {
	case s <= 8:
		return 7.0
	case s <= 12:
		return 6.0
	case s <= 16:
		return 4.0
	case s <= 20:
		return 3.0
	case s <= 24:
		return 2.0
	}
	return 1.0
}

func addDataPoints(metric *datastore.OTelMetricEnt, m *pmetric.Metric) {
	t := m.Type().String()
	switch t {
	case "Histogram":
		for l, h := range m.Histogram().DataPoints().All() {
			dp := &datastore.OTelMetricDataPointEnt{
				Start:          h.StartTimestamp().AsTime().UnixNano(),
				Time:           h.Timestamp().AsTime().UnixNano(),
				Attributes:     getAttributes(h.Attributes().AsRaw()),
				Count:          h.Count(),
				BucketCounts:   h.BucketCounts().AsRaw(),
				ExplicitBounds: h.ExplicitBounds().AsRaw(),
				Index:          l,
			}
			if h.HasSum() {
				dp.Sum = h.Sum()
			}
			if h.HasMin() {
				dp.Min = h.Min()
			}
			if h.HasMin() {
				dp.Max = h.Max()
			}
			metric.DataPoints = append(metric.DataPoints, dp)
		}
	case "Sum":
		for l, s := range m.Sum().DataPoints().All() {
			dp := &datastore.OTelMetricDataPointEnt{
				Start:      s.StartTimestamp().AsTime().UnixNano(),
				Time:       s.Timestamp().AsTime().UnixNano(),
				Attributes: getAttributes(s.Attributes().AsRaw()),
				Index:      l,
			}
			switch s.ValueType().String() {
			case "Double":
				dp.Sum = s.DoubleValue()
			case "Int":
				dp.Sum = float64(s.IntValue())
			}
			metric.DataPoints = append(metric.DataPoints, dp)
		}
	case "Gauge":
		for l, g := range m.Gauge().DataPoints().All() {
			dp := &datastore.OTelMetricDataPointEnt{
				Start:      g.StartTimestamp().AsTime().UnixNano(),
				Time:       g.Timestamp().AsTime().UnixNano(),
				Attributes: getAttributes(g.Attributes().AsRaw()),
				Index:      l,
			}
			switch g.ValueType().String() {
			case "Double":
				dp.Gauge = g.DoubleValue()
			case "Int":
				dp.Gauge = float64(g.IntValue())
			}
			metric.DataPoints = append(metric.DataPoints, dp)
		}
	case "ExponentialHistogram":
		for l, eh := range m.ExponentialHistogram().DataPoints().All() {
			dp := &datastore.OTelMetricDataPointEnt{
				Start:      eh.StartTimestamp().AsTime().UnixNano(),
				Time:       eh.Timestamp().AsTime().UnixNano(),
				Attributes: getAttributes(eh.Attributes().AsRaw()),
				Count:      eh.Count(),
				Positive:   eh.Positive().BucketCounts().AsRaw(),
				Negative:   eh.Negative().BucketCounts().AsRaw(),
				Index:      l,
			}
			if eh.HasSum() {
				dp.Sum = eh.Sum()
			}
			if eh.HasMin() {
				dp.Min = eh.Min()
			}
			if eh.HasMin() {
				dp.Max = eh.Max()
			}
			metric.DataPoints = append(metric.DataPoints, dp)
		}
	default:
		log.Printf("uknown otel metric type %s %s %s %s", m.Name(), m.Type().String(), m.Unit(), m.Description())
	}
	if len(metric.DataPoints) > 1000 {
		for i, dp := range metric.DataPoints {
			if i != dp.Index {
				metric.DataPoints = slices.Delete(metric.DataPoints, 0, i)
				break
			}
		}
	}
}

func getAttributes(m map[string]any) []string {
	ret := []string{}
	for k, v := range m {
		ret = append(ret, fmt.Sprintf("%s=%v", k, v))
	}
	sort.Strings(ret)
	return ret
}
