package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
)

func tcpd(stopCh chan bool) {
	log.Printf("start tcpd")
	sv, err := net.Listen("tcp", tcpListen)
	if err != nil {
		log.Printf("tcpd err=%v", err)
		<-stopCh
		return
	}
	defer sv.Close()
	connList := []net.Conn{}
	go func() {
		for {
			conn, err := sv.Accept()
			if err != nil {
				log.Printf("tcpd err=%v", err)
				break
			} else {
				connList = append(connList, conn)
				go tcpClientProcess(conn)
			}
		}
	}()
	<-stopCh
	for _, c := range connList {
		if err := c.Close(); err != nil {
			log.Printf("tcpd err=%v", err)
		}
	}
	log.Printf("stop tcpd")
}

func tcpClientProcess(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("tcpd recovered from panic: %v", r)
			datastore.SetPanic(fmt.Sprintf("tcpd panic=%v", r))
		}
		conn.Close()
	}()
	buf := make([]byte, 1024*1024)
	ra, ok := conn.RemoteAddr().(*net.TCPAddr)
	if !ok {
		log.Printf("tcpd no from IP")
		return
	}
	log.Printf("tcpd connect from %s", ra.String())
	host := ra.IP.String()
	if n := datastore.FindNodeFromIP(host); n != nil {
		host = n.Name
	}
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("tcpd err=%v:", err)
			break
		}
		c := 0
		for _, l := range strings.Split(string(buf[:n]), "\n") {
			if len(l) < 2 {
				continue
			}
			ts := time.Now().UnixNano()
			sl := make(map[string]interface{})
			sl["hostname"] = host
			sl["tag"] = "tcpd"
			sl["content"] = l
			sl["client"] = conn.RemoteAddr().String()
			sl["facility"] = 23
			sl["severity"] = 6
			sl["priority"] = 190
			sl["timestamp"] = time.Unix(0, ts).Format(time.RFC3339)
			s, err := json.Marshal(sl)
			if err == nil {
				c++
				logCh <- &datastore.LogEnt{
					Time: ts,
					Type: "syslog",
					Log:  string(s),
				}
			}
		}
		report.UpdateSensor(host, "tcpd", c)
	}
}
