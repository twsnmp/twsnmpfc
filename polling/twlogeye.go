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

func doPollingTwLogEye(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		log.Printf("node not found id=%x", pe.NodeID)
		return
	}
	client, err := getTwLogEyeClient(n, pe)
	if err != nil {
		setPollingError("twlogeye", pe, err)
		log.Printf("getTwLogEyeClient err=%v", err)
		return
	}
	var regFilter *regexp.Regexp
	if pe.Filter != "" {
		if regFilter, err = regexp.Compile(pe.Filter); err != nil {
			log.Printf("filter compile err=%v", err)
			setPollingError("twlogeye", pe, err)
			return
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
		log.Printf("twLogEye search notify err=%v", err)
		return
	}
	count := 0
	l := ""
	for {
		r, err := s.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Printf("search notify err=%v", err)
			setPollingError("twlogeye", pe, err)
			return
		}
		l = fmt.Sprintf("%s %s %s %s %s %s", getTimeStr(r.GetTime()), r.GetSrc(), r.GetLevel(), r.GetId(), r.GetTags(), r.GetTitle())
		log.Println(l)
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
		return
	}
	vm := otto.New()
	addJavaScriptFunctions(pe, vm)
	vm.Set("count", count)
	vm.Set("interval", pe.PollInt)
	value, err := vm.Run(pe.Script)
	if err != nil {
		setPollingError("twlogeye", pe, fmt.Errorf("invalid script err=%v", err))
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func getTwLogEyeClient(n *datastore.NodeEnt, pe *datastore.PollingEnt) (api.TWLogEyeServiceClient, error) {
	var conn *grpc.ClientConn
	var err error
	port := 8081
	if i, err := strconv.Atoi(pe.Params); err == nil {
		port = i
	}
	address := fmt.Sprintf("%s:%d", n.IP, port)
	if datastore.CACert == "" {
		// not TLS
		conn, err = grpc.NewClient(
			address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, err
		}
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
			conn, err = grpc.NewClient(address, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
			if err != nil {
				return nil, err
			}
		} else {
			// TLS
			creds, err := credentials.NewClientTLSFromFile(datastore.CACert, "")
			if err != nil {
				return nil, err
			}
			conn, err = grpc.NewClient(address, grpc.WithTransportCredentials(creds))
			if err != nil {
				return nil, err
			}
		}
	}
	return api.NewTWLogEyeServiceClient(conn), nil
}

func getTimeStr(t int64) string {
	return time.Unix(0, t).Format(time.RFC3339Nano)
}
