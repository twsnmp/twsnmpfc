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
)

var (
	pingSendCh chan *PingEnt
	randGen    *rand.Rand
	pingMutex  sync.Mutex
)

type PingEnt struct {
	Target   string
	Router   string
	Timeout  int
	Retry    int
	Size     int
	ipaddr   *net.IPAddr
	id       int
	sequence int
	Tracker  int64
	Stat     PingStat
	Time     int64
	lastSend int64
	Error    error
	done     chan bool
}

type packet struct {
	bytes  []byte
	nbytes int
	ttl    int
}

func StartPing(ctx context.Context) error {
	pingSendCh = make(chan *PingEnt, 100)
	randGen = rand.New(rand.NewSource(time.Now().UnixNano()))
	go pingBackend(ctx)
	return nil
}

// DoPing : pingの実行
func DoPing(ip string, timeout, retry, size int) *PingEnt {
	var err error
	var pe = newPingEnt(ip, timeout, retry, size)
	if pe.ipaddr, err = net.ResolveIPAddr("ip", ip); err != nil {
		pe.Stat = PingOtherError
		return pe
	}
	pingSendCh <- pe
	<-pe.done
	return pe
}

func newPingEnt(ip string, timeout, retry, size int) *PingEnt {
	pingMutex.Lock()
	defer pingMutex.Unlock()
	return &PingEnt{
		Target:   ip,
		Stat:     PingStart,
		Timeout:  timeout,
		Retry:    retry,
		Size:     size,
		sequence: 0,
		id:       randGen.Intn(math.MaxInt16),
		Tracker:  randGen.Int63n(math.MaxInt64),
		done:     make(chan bool),
	}
}

func (p *PingEnt) sendICMP(conn *icmp.PacketConn) error {
	p.lastSend = time.Now().Unix()
	var dst net.Addr = p.ipaddr
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		dst = &net.UDPAddr{IP: p.ipaddr.IP, Zone: p.ipaddr.Zone}
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
func pingBackend(ctx context.Context) {
	timer := time.NewTicker(time.Millisecond * 500)
	pingMap := make(map[int64]*PingEnt)
	netProto := "udp4"
	if runtime.GOOS == "windows" {
		netProto = "ip4:icmp"
	}
	conn, err := icmp.ListenPacket(netProto, "0.0.0.0")
	if err != nil {
		log.Printf("pingBackend err=%v", err)
		return
	}
	defer conn.Close()
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			for _, p := range pingMap {
				close(p.done)
			}
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
					log.Printf("sendICMP err=%v", err)
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
						log.Printf("sendICMP err=%v", err)
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
				log.Printf("pingBackend err=%v", err)
				continue
			}
			if tracker, tm, err := processPacket(&packet{bytes: bytes, nbytes: n, ttl: ttl}); err != nil {
				log.Printf("pingBackend processPacket err=%v", err)
			} else {
				if p, ok := pingMap[tracker]; ok {
					sa := strings.Split(src.String(), ":")
					if p.Target != sa[0] {
						log.Printf("pingBackend target=%s src=%s", p.Target, src.String())
						continue
					}
					delete(pingMap, tracker)
					p.Stat = PingOK
					p.Time = tm
					p.Error = nil
					p.done <- true
				}
			}
		}
	}
}

func processPacket(recv *packet) (int64, int64, error) {
	receivedAt := time.Now()
	var m *icmp.Message
	var err error
	if m, err = icmp.ParseMessage(protocolICMP, recv.bytes); err != nil {
		return -1, -1, fmt.Errorf("error parsing icmp message: %s", err.Error())
	}
	if m.Type != ipv4.ICMPTypeEchoReply {
		return -1, -1, fmt.Errorf("icmp message type != ICMPTypeEchoReply  : %v", m)
	}
	switch pkt := m.Body.(type) {
	case *icmp.Echo:
		if len(pkt.Data) < timeSliceLength+trackerLength {
			return -1, -1, fmt.Errorf("insufficient data received; got: %d %v", len(pkt.Data), pkt.Data)
		}
		tracker := bytesToInt(pkt.Data[timeSliceLength:])
		timestamp := bytesToTime(pkt.Data[:timeSliceLength])
		return tracker, receivedAt.Sub(timestamp).Nanoseconds(), nil
	default:
		// Very bad, not sure how this can happen
		return -1, -1, fmt.Errorf("invalid ICMP echo reply; type: '%T', '%v'", pkt, pkt)
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
