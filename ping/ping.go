// Package ping : pingの実行
package ping

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	timeSliceLength = 8
	trackerLength   = 8
	protocolICMP    = 1
)

type PingStat int

const (
	PingStart = iota
	PingOK
	PingTimeout
	PingOtherError
	PingTimeExceeded
)

var (
	pingSendCh chan *PingEnt
	randGen    *rand.Rand
	pingMutex  sync.Mutex
	pingMode   = ""
)

type PingEnt struct {
	Target   string
	Router   string
	Timeout  int
	Retry    int
	Size     int
	SendTTL  int
	RecvTTL  int
	ipaddr   *net.IPAddr
	id       int
	sequence int
	Tracker  int64
	Stat     PingStat
	Time     int64
	lastSend int64
	Error    error
	RecvSrc  string
	done     chan bool
}

type packet struct {
	bytes  []byte
	nbytes int
	ttl    int
}

func Start(ctx context.Context, wg *sync.WaitGroup, mode string) error {
	if mode == "" {
		if runtime.GOOS == "darwin" {
			mode = "udp"
		} else {
			mode = "icmp"
		}
	}
	pingMode = mode
	log.Printf("ping mode=%s", pingMode)
	pingSendCh = make(chan *PingEnt, 100)
	randGen = rand.New(rand.NewSource(time.Now().UnixNano()))
	wg.Add(1)
	go pingBackend(ctx, wg)
	return nil
}

// DoPing : pingの実行
func DoPing(ip string, timeout, retry, size, ttl int) *PingEnt {
	var err error
	var pe = newPingEnt(ip, timeout, retry, size, ttl)
	if pe.ipaddr, err = net.ResolveIPAddr("ip", ip); err != nil {
		pe.Stat = PingOtherError
		return pe
	}
	pingSendCh <- pe
	<-pe.done
	return pe
}

func newPingEnt(ip string, timeout, retry, size, ttl int) *PingEnt {
	pingMutex.Lock()
	defer pingMutex.Unlock()
	return &PingEnt{
		Target:   ip,
		Stat:     PingStart,
		Timeout:  timeout,
		Retry:    retry,
		Size:     size,
		SendTTL:  ttl,
		sequence: 0,
		id:       randGen.Intn(math.MaxInt16),
		Tracker:  randGen.Int63n(math.MaxInt64),
		done:     make(chan bool),
	}
}

func (p *PingEnt) sendICMP(conn *icmp.PacketConn) error {
	p.lastSend = time.Now().Unix()
	var dst net.Addr = p.ipaddr
	if pingMode == "udp" {
		dst = &net.UDPAddr{IP: p.ipaddr.IP, Zone: p.ipaddr.Zone}
	}
	ipcon := conn.IPv4PacketConn()
	if ipcon != nil && p.SendTTL > 0 && p.SendTTL < 256 {
		ipcon.SetTTL(p.SendTTL)
	}
	t := append(timeToBytes(time.Now()), intToBytes(p.Tracker)...)
	if remainSize := p.Size - timeSliceLength - trackerLength; remainSize > 0 {
		t = append(t, bytes.Repeat([]byte{1}, remainSize)...)
	}

	body := &icmp.Echo{
		ID:   p.id,
		Seq:  p.sequence,
		Data: t,
	}

	msg := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: body,
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return err
	}
	for {
		if _, err := conn.WriteTo(msgBytes, dst); err != nil {
			if neterr, ok := err.(*net.OpError); ok {
				if neterr.Err == syscall.ENOBUFS {
					continue
				}
			}
			return err
		}
		break
	}
	return nil
}

// pingBackend : ping実行時の送受信処理
func pingBackend(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start ping")
	timer := time.NewTicker(time.Millisecond * 500)
	pingMap := make(map[int64]*PingEnt)
	netProto := "ip4:icmp"
	if pingMode == "udp" {
		netProto = "udp4"
	}
	conn, err := icmp.ListenPacket(netProto, "0.0.0.0")
	if err != nil {
		log.Fatalf("ping listen err=%v", err)
		return
	}
	defer conn.Close()
	conn.IPv4PacketConn().SetControlMessage(ipv4.FlagTTL, true)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			for _, p := range pingMap {
				close(p.done)
			}
			log.Println("stop ping")
			return
		case p := <-pingSendCh:
			if p != nil {
				_, ok := pingMap[p.Tracker]
				for ok {
					p.Tracker++
					_, ok = pingMap[p.Tracker]
				}
				pingMap[p.Tracker] = p
				if err := p.sendICMP(conn); err != nil {
					p.Error = err
				}
			}
		case <-timer.C:
			now := time.Now().Unix()
			for k, p := range pingMap {
				if p.lastSend+int64(p.Timeout) < now {
					p.sequence++
					if p.sequence > p.Retry {
						delete(pingMap, k)
						if p.Error == nil {
							p.Error = fmt.Errorf("Timeout")
						}
						p.Stat = PingTimeout
						p.done <- true
						continue
					}
					if err := p.sendICMP(conn); err != nil {
						p.Error = err
						log.Printf("ping send err=%v", err)
					}
				}
			}
		default:
			bytes := make([]byte, 2048)
			_ = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
			var n, ttl int
			var err error
			var cm *ipv4.ControlMessage
			var src net.Addr
			n, cm, src, err = conn.IPv4PacketConn().ReadFrom(bytes)
			if cm != nil {
				ttl = cm.TTL
			}
			if err != nil {
				if neterr, ok := err.(*net.OpError); ok {
					if neterr.Timeout() {
						// Read timeout
						continue
					}
				}
				continue
			}
			if tracker, tm, te, err := processPacket(&packet{bytes: bytes, nbytes: n, ttl: ttl}); err == nil {
				if p, ok := pingMap[tracker]; ok {
					sa := strings.Split(src.String(), ":")
					if p.Target != sa[0] && !te {
						log.Printf("ping target=%s src=%s", p.Target, src.String())
						continue
					}
					delete(pingMap, tracker)
					if te {
						p.Stat = PingTimeExceeded
					} else {
						p.Stat = PingOK
					}
					p.Time = tm
					p.RecvTTL = ttl
					p.RecvSrc = sa[0]
					p.Error = nil
					p.done <- true
				}
			}
		}
	}
}

func processIcmpTimeExceeded(b []byte) (int64, int64, bool, error) {
	iph, err := ipv4.ParseHeader(b)
	if err != nil {
		log.Println(err)
		return -1, -1, false, err
	}
	if iph.Len+timeSliceLength+8 > len(b) {
		return -1, -1, false, fmt.Errorf("icmp time exceeded legth error")
	}
	var m *icmp.Message
	if m, err = icmp.ParseMessage(protocolICMP, b[iph.Len:]); err != nil {
		return -1, -1, false, fmt.Errorf("error parsing icmp message in timeexceeded : %v", err)
	}
	if m.Type != ipv4.ICMPTypeEcho {
		log.Printf("not echo %v", m)
		return -1, -1, false, fmt.Errorf("icmp message type in timeexedded error type=%v", m)
	}
	switch pkt := m.Body.(type) {
	case *icmp.Echo:
		if len(pkt.Data) < timeSliceLength+trackerLength {
			return -1, -1, false, fmt.Errorf("insufficient data received; got: %d %v", len(pkt.Data), pkt.Data)
		}
		receivedAt := time.Now()
		tracker := bytesToInt(pkt.Data[timeSliceLength:])
		timestamp := bytesToTime(pkt.Data[:timeSliceLength])
		return tracker, receivedAt.Sub(timestamp).Nanoseconds(), true, nil
	}
	return -1, -1, false, fmt.Errorf("not icmp echo")
}

func processPacket(recv *packet) (int64, int64, bool, error) {
	receivedAt := time.Now()
	var m *icmp.Message
	var err error
	if m, err = icmp.ParseMessage(protocolICMP, recv.bytes); err != nil {
		return -1, -1, false, fmt.Errorf("error parsing icmp message: %s", err.Error())
	}
	if m.Type != ipv4.ICMPTypeEchoReply && m.Type != ipv4.ICMPTypeTimeExceeded {
		return -1, -1, false, fmt.Errorf("icmp message type error type=%v", m)
	}
	switch pkt := m.Body.(type) {
	case *icmp.Echo:
		if len(pkt.Data) < timeSliceLength+trackerLength {
			return -1, -1, false, fmt.Errorf("insufficient data received; got: %d %v", len(pkt.Data), pkt.Data)
		}
		tracker := bytesToInt(pkt.Data[timeSliceLength:])
		timestamp := bytesToTime(pkt.Data[:timeSliceLength])
		return tracker, receivedAt.Sub(timestamp).Nanoseconds(), false, nil
	case *icmp.TimeExceeded:
		return processIcmpTimeExceeded(pkt.Data)
	default:
		// Very bad, not sure how this can happen
		return -1, -1, false, fmt.Errorf("invalid icmp echo reply; type: '%T', '%v'", pkt, pkt)
	}
}

func bytesToTime(b []byte) time.Time {
	var nsec int64
	for i := uint8(0); i < 8; i++ {
		nsec += int64(b[i]) << ((7 - i) * 8)
	}
	return time.Unix(nsec/1000000000, nsec%1000000000)
}

func timeToBytes(t time.Time) []byte {
	nsec := t.UnixNano()
	b := make([]byte, 8)
	for i := uint8(0); i < 8; i++ {
		b[i] = byte((nsec >> ((7 - i) * 8)) & 0xff)
	}
	return b
}

func bytesToInt(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

func intToBytes(tracker int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(tracker))
	return b
}
