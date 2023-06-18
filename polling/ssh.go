package polling

// SSH コマンド実行で監視する。

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/vjeantet/grok"
	"golang.org/x/crypto/ssh"
)

func doPollingSSH(pe *datastore.PollingEnt) {
	cmd := pe.Params
	extractor := pe.Extractor
	script := pe.Script
	port := "22"
	if pe.Mode != "" {
		port = pe.Mode
	}
	vm := otto.New()
	addJavaScriptFunctions(pe, vm)
	cl := strings.Split(cmd, " ")
	if len(cl) < 1 {
		setPollingError("ssh", pe, fmt.Errorf("no cmd"))
		return
	}
	client, session, err := sshConnectToHost(pe, port)
	if err != nil {
		setPollingError("ssh", pe, err)
		return
	}
	defer func() {
		session.Close()
		client.Close()
	}()
	exitCode := 0
	out, err := session.CombinedOutput(cmd)
	if err != nil {
		if e, ok := err.(*ssh.ExitError); ok {
			exitCode = e.Waitmsg.ExitStatus()
		} else {
			setPollingError("ssh", pe, err)
			return
		}
	}
	pe.Result["lastTime"] = time.Now().Format("2006-01-02T15:04")
	pe.Result["exitCode"] = float64(exitCode)
	pe.Result["error"] = ""
	vm.Set("interval", pe.PollInt)
	vm.Set("exitCode", exitCode)
	if extractor == "goquery" {
		setPollingError("ssh", pe, fmt.Errorf("goquery not supported"))
		return
	} else if extractor == "getBody" {
		vm.Set("getBody", func(call otto.FunctionCall) otto.Value {
			if r, err := otto.ToValue(string(out)); err == nil {
				return r
			}
			return otto.UndefinedValue()
		})
	} else if extractor != "" {
		grokEnt := datastore.GetGrokEnt(extractor)
		if grokEnt == nil {
			setPollingError("ssh", pe, fmt.Errorf("no grok pattern"))
			return
		}
		g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
		g.AddPattern(extractor, grokEnt.Pat)
		cap := fmt.Sprintf("%%{%s}", extractor)
		values, err := g.Parse(cap, string(out))
		if err != nil {
			setPollingError("ssh", pe, err)
			return
		}
		for k, v := range values {
			vm.Set(k, v)
			pe.Result[k] = v
		}
	}
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("ssh", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
		return
	}
	setPollingState(pe, pe.Level)
}

func sshConnectToHost(pe *datastore.PollingEnt, port string) (*ssh.Client, *ssh.Session, error) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		return nil, nil, fmt.Errorf("node not found nodeID=%s", pe.NodeID)
	}
	signer, err := ssh.ParsePrivateKey([]byte(datastore.GetPrivateKey()))
	if err != nil {
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
				return nil
			}
		//ssh.InsecureIgnoreHostKey()
	}
	conn, err := net.DialTimeout("tcp", n.IP+":"+port, time.Duration(pe.Timeout)*time.Second)
	if err != nil {
		return nil, nil, err
	}
	if err := conn.SetDeadline(time.Now().Add(time.Second * time.Duration(pe.PollInt-5))); err != nil {
		return nil, nil, err
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
