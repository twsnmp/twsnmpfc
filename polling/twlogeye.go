package polling

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/robertkrimen/otto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/twsnmp/twlogeye/api"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func doPollingTwLogEye(pe *datastore.PollingEnt) bool {
	switch pe.Mode {
	case "report.syslog":
		return doPollingTwLogEyeSyslogReport(pe)
	case "report.trap", "report.snmptrap":
		return doPollingTwLogEyeSnmpTrapReport(pe)
	case "report.netflow":
		return doPollingTwLogEyeNetflowReport(pe)
	case "report.winevent":
		return doPollingTwLogEyeWindowsEventReport(pe)
	case "report.anomaly":
		return doPollingTwLogEyeAnomalyReport(pe)
	default: // notify
		return doPollingTwLogEyeNotify(pe)
	}
}

func doPollingTwLogEyeNotify(pe *datastore.PollingEnt) bool {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		log.Printf("twlogeye polling node not found id=%x", pe.NodeID)
		return false
	}
	conn, err := getTwLogEyeClientConn(n, pe)
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	defer conn.Close()
	client := api.NewTWLogEyeServiceClient(conn)
	var regFilter *regexp.Regexp
	if pe.Filter != "" {
		if regFilter, err = regexp.Compile(pe.Filter); err != nil {
			setPollingError("twlogeye", pe, err)
			return false
		}
	}
	st := time.Now().Add(time.Duration(pe.PollInt) * time.Second * -1).UnixNano()
	et := time.Now().UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if lt, ok := v.(int64); ok {
			st = lt
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pe.Timeout)*time.Second)
	defer cancel()
	s, err := client.SearchNotify(ctx, &api.NofifyRequest{Level: pe.Mode, Start: st, End: et})
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	count := 0
	l := ""
	for {
		r, err := s.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			setPollingError("twlogeye", pe, err)
			return false
		}
		l = fmt.Sprintf("%s %s %s %s %s %s", getTimeStr(r.GetTime()), r.GetSrc(), r.GetLevel(), r.GetId(), r.GetTags(), r.GetTitle())
		if regFilter != nil && !regFilter.MatchString(l) {
			continue
		}
		count++
	}
	pe.Result["lastTime"] = et
	pe.Result["count"] = float64(count)
	if count > 0 {
		pe.Result["lastLog"] = l
	}
	if pe.Script == "" {
		setPollingState(pe, "normal")
		return true
	}
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	vm.Set("count", count)
	vm.Set("interval", pe.PollInt)
	value, err := vm.Run(pe.Script)
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
	return true
}

func doPollingTwLogEyeSyslogReport(pe *datastore.PollingEnt) bool {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		log.Printf("node not found id=%x", pe.NodeID)
		return false
	}
	conn, err := getTwLogEyeClientConn(n, pe)
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	defer conn.Close()
	client := api.NewTWLogEyeServiceClient(conn)
	st := time.Now().Add(time.Duration(pe.PollInt) * time.Second * -1).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if lt, ok := v.(int64); ok {
			st = lt
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pe.Timeout+4)*time.Second)
	defer cancel()
	l, err := client.GetLastSyslogReport(ctx, &api.Empty{})
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	if st >= l.Time {
		return false
	}
	pe.Result["lastTime"] = l.Time
	pe.Result["errors"] = float64(l.GetError())
	pe.Result["warns"] = float64(l.GetWarn())
	pe.Result["normal"] = float64(l.GetNormal())
	pe.Result["patterns"] = float64(l.GetPatterns())
	pe.Result["errPatterns"] = float64(l.GetErrPatterns())
	if pe.Script == "" {
		setPollingState(pe, "normal")
		return true
	}
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	vm.Set("interval", pe.PollInt)
	value, err := vm.Run(pe.Script)
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
	return true
}

func doPollingTwLogEyeSnmpTrapReport(pe *datastore.PollingEnt) bool {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		log.Printf("node not found id=%x", pe.NodeID)
		return false
	}
	conn, err := getTwLogEyeClientConn(n, pe)
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	defer conn.Close()
	client := api.NewTWLogEyeServiceClient(conn)
	st := time.Now().Add(time.Duration(pe.PollInt) * time.Second * -1).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if lt, ok := v.(int64); ok {
			st = lt
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pe.Timeout)*time.Second)
	defer cancel()
	l, err := client.GetLastTrapReport(ctx, &api.Empty{})
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	if st >= l.Time {
		return false
	}
	pe.Result["lastTime"] = l.Time
	pe.Result["count"] = float64(l.GetCount())
	pe.Result["types"] = float64(l.GetTypes())
	if pe.Script == "" {
		setPollingState(pe, "normal")
		return true
	}
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	vm.Set("interval", pe.PollInt)
	value, err := vm.Run(pe.Script)
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
	return true
}

func doPollingTwLogEyeNetflowReport(pe *datastore.PollingEnt) bool {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		log.Printf("node not found id=%x", pe.NodeID)
		return false
	}
	conn, err := getTwLogEyeClientConn(n, pe)
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	defer conn.Close()
	client := api.NewTWLogEyeServiceClient(conn)
	st := time.Now().Add(time.Duration(pe.PollInt) * time.Second * -1).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if lt, ok := v.(int64); ok {
			st = lt
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pe.Timeout)*time.Second)
	defer cancel()
	l, err := client.GetLastNetflowReport(ctx, &api.Empty{})
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	if st >= l.Time {
		return false
	}
	pe.Result["lastTime"] = l.Time
	pe.Result["bytes"] = float64(l.GetBytes())
	pe.Result["flows"] = float64(l.GetFlows())
	pe.Result["fumbles"] = float64(l.GetFumbles())
	pe.Result["IPs"] = float64(l.GetIps())
	pe.Result["MACs"] = float64(l.GetMacs())
	pe.Result["protocols"] = float64(l.GetProtocols())
	pe.Result["packets"] = float64(l.GetPackets())
	if pe.Script == "" {
		setPollingState(pe, "normal")
		return true
	}
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	vm.Set("interval", pe.PollInt)
	value, err := vm.Run(pe.Script)
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
	return true
}

func doPollingTwLogEyeWindowsEventReport(pe *datastore.PollingEnt) bool {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		log.Printf("node not found id=%x", pe.NodeID)
		return false
	}
	conn, err := getTwLogEyeClientConn(n, pe)
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	defer conn.Close()
	client := api.NewTWLogEyeServiceClient(conn)
	st := time.Now().Add(time.Duration(pe.PollInt) * time.Second * -1).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if lt, ok := v.(int64); ok {
			st = lt
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pe.Timeout)*time.Second)
	defer cancel()
	l, err := client.GetLastWindowsEventReport(ctx, &api.Empty{})
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	if st >= l.Time {
		return false
	}
	pe.Result["lastTime"] = l.Time
	pe.Result["errors"] = float64(l.GetError())
	pe.Result["warns"] = float64(l.GetWarn())
	pe.Result["normal"] = float64(l.GetNormal())
	pe.Result["types"] = float64(l.GetTypes())
	pe.Result["errTypes"] = float64(l.GetErrorTypes())
	if pe.Script == "" {
		setPollingState(pe, "normal")
		return true
	}
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	vm.Set("interval", pe.PollInt)
	value, err := vm.Run(pe.Script)
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
	return true
}

func doPollingTwLogEyeAnomalyReport(pe *datastore.PollingEnt) bool {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		log.Printf("node not found id=%x", pe.NodeID)
		return false
	}
	conn, err := getTwLogEyeClientConn(n, pe)
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	defer conn.Close()
	client := api.NewTWLogEyeServiceClient(conn)
	st := time.Now().Add(time.Duration(pe.PollInt) * time.Second * -1).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if lt, ok := v.(int64); ok {
			st = lt
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pe.Timeout)*time.Second)
	defer cancel()
	l, err := client.GetLastAnomalyReport(ctx, &api.Empty{})
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	if st >= l.Time {
		return false
	}
	pe.Result["lastTime"] = l.Time
	for _, e := range l.GetScoreList() {
		pe.Result[e.Type+"_Score"] = e.GetScore()
	}
	if pe.Script == "" {
		setPollingState(pe, "normal")
		return true
	}
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	vm.Set("interval", pe.PollInt)
	value, err := vm.Run(pe.Script)
	if err != nil {
		setPollingError("twlogeye", pe, err)
		return false
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
	return true
}

func getTwLogEyeClientConn(n *datastore.NodeEnt, pe *datastore.PollingEnt) (*grpc.ClientConn, error) {
	port := 8081
	if i, err := strconv.Atoi(pe.Params); err == nil {
		port = i
	}
	address := fmt.Sprintf("%s:%d", n.IP, port)
	if datastore.CACert == "" {
		// not TLS
		return grpc.NewClient(
			address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	} else {
		if datastore.ClientCert != "" && datastore.ClientKey != "" {
			// mTLS
			cert, err := tls.LoadX509KeyPair(datastore.ClientCert, datastore.ClientKey)
			if err != nil {
				return nil, err
			}
			ca := x509.NewCertPool()
			caBytes, err := os.ReadFile(datastore.CACert)
			if err != nil {
				return nil, err
			}
			if ok := ca.AppendCertsFromPEM(caBytes); !ok {
				return nil, err
			}
			tlsConfig := &tls.Config{
				ServerName:   "",
				Certificates: []tls.Certificate{cert},
				RootCAs:      ca,
			}
			return grpc.NewClient(address, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
		} else {
			// TLS
			creds, err := credentials.NewClientTLSFromFile(datastore.CACert, "")
			if err != nil {
				return nil, err
			}
			return grpc.NewClient(address, grpc.WithTransportCredentials(creds))
		}
	}
}

func getTimeStr(t int64) string {
	return time.Unix(0, t).Format(time.RFC3339Nano)
}
