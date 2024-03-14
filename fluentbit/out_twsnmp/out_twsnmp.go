package main

import (
	"C"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"golang.org/x/crypto/ssh"
)
import (
	"io"
	"net"
	"strconv"
)

type outputConfEnt struct {
	Twsnmp      string
	PrivateKey  []byte
	HostKey     []byte
	LastHostKey string
	Facility    int
	Severity    int
	Host        string
}

var outputConfMap = make(map[string]*outputConfEnt)

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	log.Printf("[out_twsnmp] Register called")
	return output.FLBPluginRegister(def, "twsnmp", "Output plugin for TWSNMP")
}

//export FLBPluginInit
func FLBPluginInit(plugin unsafe.Pointer) int {
	id := output.FLBPluginConfigKey(plugin, "id")
	log.Printf("[out_twsnmp] id = %q", id)
	if id == "" {
		return output.FLB_ERROR
	}
	// Set the context to point to any Go variable
	output.FLBPluginSetContext(plugin, id)
	twsnmp := output.FLBPluginConfigKey(plugin, "twsnmp")
	privateKeyPath := replaceHomeDir(output.FLBPluginConfigKey(plugin, "private_key"))
	if privateKeyPath == "" {
		privateKeyPath = replaceHomeDir("~/.ssh/id_rsa")
	}
	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Printf("[out_twsnmp] read private key err:%v", err)
		return output.FLB_ERROR
	}
	hostKeyPath := replaceHomeDir(output.FLBPluginConfigKey(plugin, "host_key"))
	hostKey := []byte{}
	if hostKeyPath != "" {
		hostKey, err = os.ReadFile(hostKeyPath)
		if err != nil {
			log.Printf("[out_twsnmp] read host key err:%v", err)
			return output.FLB_ERROR
		}
	}
	fac := 21
	sev := 6
	if v, err := strconv.ParseInt(output.FLBPluginConfigKey(plugin, "facility"), 10, 64); err == nil {
		fac = int(v)
	}
	if v, err := strconv.ParseInt(output.FLBPluginConfigKey(plugin, "severity"), 10, 64); err == nil {
		sev = int(v)
	}
	host := output.FLBPluginConfigKey(plugin, "host_name")
	log.Printf("[out_twsnmp] twsnmp= '%s'", twsnmp)
	log.Printf("[out_twsnmp] privateKeyPath= '%s'", privateKeyPath)
	log.Printf("[out_twsnmp] hostKeyPath= '%s'", hostKeyPath)
	log.Printf("[out_twsnmp] facility=%d  severity=%d", fac, sev)
	log.Printf("[out_twsnmp] host= '%s'", host)

	conf := &outputConfEnt{
		Twsnmp:     twsnmp,
		PrivateKey: privateKey,
		HostKey:    hostKey,
		Facility:   fac,
		Severity:   sev,
		Host:       host,
	}
	// Check connect
	c, s, err := connectSSH(conf)
	if err != nil {
		log.Printf("[out_twsnmp] connectSSH err:%v", err)
		return output.FLB_ERROR
	}
	s.Close()
	c.Close()
	outputConfMap[id] = conf
	return output.FLB_OK
}

func replaceHomeDir(p string) string {
	if !strings.HasPrefix(p, "~") {
		return p
	}
	usr, _ := user.Current()
	return strings.Replace(p, "~", usr.HomeDir, 1)
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	log.Print("[out_twsnmp] Flush called for unknown instance")
	return output.FLB_OK
}

//export FLBPluginFlushCtx
func FLBPluginFlushCtx(ctx, data unsafe.Pointer, length C.int, tag *C.char) int {
	// Type assert context back into the original type for the Go variable
	id := output.FLBPluginGetContext(ctx).(string)
	conf, ok := outputConfMap[id]
	if !ok {
		log.Printf("[out_twsnmp] Flush called for inkown id: %s", id)
		return output.FLB_ERROR
	}

	c, s, err := connectSSH(conf)
	if err != nil {
		log.Printf("[out_twsnmp] connectSSH err: %v", err)
		return output.FLB_ERROR
	}
	defer s.Close()
	defer c.Close()
	cmd := fmt.Sprintf("put syslog %d %d", conf.Facility, conf.Severity)
	if conf.Host != "" {
		cmd += " " + conf.Host
	}
	out, err := s.StdinPipe()
	if err != nil {
		log.Printf("[out_twsnmp] ssh.StdinPipe err: %v", err)
		return output.FLB_ERROR
	}
	err = s.Start(cmd)
	if err != nil {
		log.Printf("[out_twsnmp] ssh.Start err: %v", err)
		return output.FLB_ERROR
	}

	dec := output.NewDecoder(data, int(length))
	count := 0
	for {
		ret, ts, record := output.GetRecord(dec)
		if ret != 0 {
			break
		}

		var timestamp time.Time
		switch t := ts.(type) {
		case output.FLBTime:
			timestamp = ts.(output.FLBTime).Time
		case uint64:
			timestamp = time.Unix(int64(t), 0)
		default:
			timestamp = time.Now()
		}
		msg := []string{}
		for k, v := range record {
			switch vv := v.(type) {
			case string:
				msg = append(msg, fmt.Sprintf(`"%v": "%s"`, k, vv))
			case []uint8:
				msg = append(msg, fmt.Sprintf(`"%v": "%s"`, k, string(vv)))
			default:
				msg = append(msg, fmt.Sprintf(`"%v": %v`, k, v))
			}
		}
		io.WriteString(out, fmt.Sprintf("%d\t%s\t{%s}\r\n", timestamp.UnixNano(), C.GoString(tag), strings.Join(msg, ", ")))
		count++
	}
	return output.FLB_OK
}

func connectSSH(conf *outputConfEnt) (*ssh.Client, *ssh.Session, error) {
	signer, err := ssh.ParsePrivateKey(conf.PrivateKey)
	if err != nil {
		return nil, nil, err
	}
	sshConfig := &ssh.ClientConfig{
		User:    "twsnmp",
		Auth:    []ssh.AuthMethod{},
		Timeout: time.Second * 1,
	}
	sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
	if len(conf.HostKey) > 0 {
		pubkey, _, _, _, err := ssh.ParseAuthorizedKey(conf.HostKey)
		if err != nil {
			return nil, nil, err
		}
		sshConfig.HostKeyCallback = ssh.FixedHostKey(pubkey)
	} else {
		sshConfig.HostKeyCallback =
			func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				newKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
				if conf.LastHostKey == "" {
					conf.LastHostKey = newKey
				}
				if conf.LastHostKey != newKey {
					return fmt.Errorf("host key changed")
				}
				return nil
			}
	}
	client, err := ssh.Dial("tcp", conf.Twsnmp, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}

//export FLBPluginExit
func FLBPluginExit() int {
	log.Print("[out_twsnmp] Exit called for unknown instance")
	return output.FLB_OK
}

//export FLBPluginExitCtx
func FLBPluginExitCtx(ctx unsafe.Pointer) int {
	// Type assert context back into the original type for the Go variable
	id := output.FLBPluginGetContext(ctx).(string)
	log.Printf("[out_twsnmp] Exit called for id: %s", id)
	return output.FLB_OK
}

//export FLBPluginUnregister
func FLBPluginUnregister(def unsafe.Pointer) {
	log.Print("[out_twsnmp] Unregister called")
	output.FLBPluginUnregister(def)
}

func main() {
}
