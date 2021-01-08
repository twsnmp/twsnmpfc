package polling

// SSH コマンド実行で監視する。

import (
	"fmt"
	"log"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/vjeantet/grok"
	"golang.org/x/crypto/ssh"
)

func (p *Polling) doPollingSSH(pe *datastore.PollingEnt) {
	cmds := splitCmd(pe.Polling)
	if len(cmds) < 3 {
		p.setPollingError("ssh", pe, fmt.Errorf("no cmd"))
		return
	}
	cmd := cmds[0]
	extractor := cmds[1]
	script := cmds[2]
	port := "22"
	if len(cmds) > 3 {
		port = cmds[3]
	}
	vm := otto.New()
	lr := make(map[string]string)
	cl := strings.Split(cmd, " ")
	if len(cl) < 1 {
		p.setPollingError("ssh", pe, fmt.Errorf("no cmd"))
		return
	}
	client, session, err := p.sshConnectToHost(pe, port)
	if err != nil {
		log.Printf("ssh error Polling=%s err=%v", pe.Polling, err)
		lr["error"] = fmt.Sprintf("%v", err)
		pe.LastResult = makeLastResult(lr)
		pe.LastVal = 0.0
		p.setPollingState(pe, pe.Level)
		return
	}
	defer func() {
		session.Close()
		client.Close()
	}()
	out, err := session.CombinedOutput(cmd)
	if err != nil {
		if e, ok := err.(*ssh.ExitError); ok {
			pe.LastVal = float64(e.Waitmsg.ExitStatus())
		} else {
			log.Printf("ssh error Polling=%s err=%v", pe.Polling, err)
			lr["error"] = fmt.Sprintf("%v", err)
			pe.LastResult = makeLastResult(lr)
			pe.LastVal = 0.0
			p.setPollingState(pe, pe.Level)
			return
		}
	} else {
		pe.LastVal = 0.0
	}
	lr["lastTime"] = time.Now().Format("2006-01-02T15:04")
	lr["exitCode"] = fmt.Sprintf("%d", int(pe.LastVal))
	_ = vm.Set("interval", pe.PollInt)
	_ = vm.Set("exitCode", int(pe.LastVal))
	if extractor != "" {
		grokEnt := p.ds.GetGrokEnt(extractor)
		if grokEnt == nil {
			log.Printf("No grok pattern Polling=%s", pe.Polling)
			p.setPollingError("ssh", pe, fmt.Errorf("no grok pattern"))
			return
		}
		g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
		_ = g.AddPattern(extractor, grokEnt.Pat)
		cap := fmt.Sprintf("%%{%s}", extractor)
		values, err := g.Parse(cap, string(out))
		if err != nil {
			p.setPollingError("ssh", pe, err)
			return
		}
		for k, v := range values {
			_ = vm.Set(k, v)
			lr[k] = v
		}
	}
	value, err := vm.Run(script)
	if err != nil {
		p.setPollingError("ssh", pe, err)
		return
	}
	pe.LastVal = 0.0
	for k, v := range lr {
		if strings.Contains(script, k) {
			if fv, err := strconv.ParseFloat(v, 64); err != nil || !math.IsNaN(fv) {
				pe.LastVal = fv
			}
			break
		}
	}
	pe.LastResult = makeLastResult(lr)
	if ok, _ := value.ToBoolean(); ok {
		p.setPollingState(pe, "normal")
		return
	}
	p.setPollingState(pe, pe.Level)
}

func (p *Polling) sshConnectToHost(pe *datastore.PollingEnt, port string) (*ssh.Client, *ssh.Session, error) {
	n := p.ds.GetNode(pe.NodeID)
	if n == nil {
		log.Printf("node not found nodeID=%s", pe.NodeID)
		return nil, nil, fmt.Errorf("node not found nodeID=%s", pe.NodeID)
	}
	signer, err := ssh.ParsePrivateKey([]byte(p.ds.GetPrivateKey()))
	if err != nil {
		log.Printf("sshConnectToHost err=%v", err)
		return nil, nil, fmt.Errorf("no private key for ssh")
	}
	sshConfig := &ssh.ClientConfig{
		User:    n.User,
		Auth:    []ssh.AuthMethod{},
		Timeout: time.Duration(pe.Timeout) * time.Second,
	}
	sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
	if n.Password != "" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.Password(n.Password))
	}
	if n.PublicKey != "" {
		pubkey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(n.PublicKey))
		if err != nil {
			return nil, nil, fmt.Errorf("invalid public key=%s", n.PublicKey)
		}
		sshConfig.HostKeyCallback = ssh.FixedHostKey(pubkey)
	} else {
		sshConfig.HostKeyCallback =
			func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				n.PublicKey = strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
				if err := p.ds.UpdateNode(n); err != nil {
					log.Printf("sshConnectToHost err=%v", err)
				}
				p.pollingStateChangeCh <- pe
				return nil
			}
		//ssh.InsecureIgnoreHostKey()
	}
	conn, err := net.DialTimeout("tcp", n.IP+":"+port, time.Duration(pe.Timeout)*time.Second)
	if err != nil {
		return nil, nil, err
	}
	if err := conn.SetDeadline(time.Now().Add(time.Second * time.Duration(pe.PollInt-5))); err != nil {
		log.Printf("sshConnectToHost err=%v", err)
	}
	c, ch, req, err := ssh.NewClientConn(conn, n.IP+":"+port, sshConfig)
	if err != nil {
		return nil, nil, err
	}
	client := ssh.NewClient(c, ch, req)
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	return client, session, nil
}
