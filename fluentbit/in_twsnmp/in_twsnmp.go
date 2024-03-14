package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/input"
	"golang.org/x/crypto/ssh"
)

// Params
var twsnmp = ""
var privateKey = []byte{}
var hostKey = []byte{}
var logType = "syslog"

// local vars
var lastTime int64
var lastHostKey = ""

type logEnt struct {
	TimeStamp int64
	Data      *map[string]interface{}
}

var logList = []*logEnt{}

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	log.Printf("[in_twsnmp] Register called")
	return input.FLBPluginRegister(def, "twsnmp", "Input plugin for TWSNMP")
}

// (fluentbit will call this)
// plugin (context) pointer to fluentbit context (state/ c code)
//
//export FLBPluginInit
func FLBPluginInit(plugin unsafe.Pointer) int {
	var err error
	twsnmp = input.FLBPluginConfigKey(plugin, "twsnmp")
	privateKeyPath := replaceHomeDir(input.FLBPluginConfigKey(plugin, "private_key"))
	if privateKeyPath == "" {
		privateKeyPath = replaceHomeDir("~/.ssh/id_rsa")
	}
	privateKey, err = os.ReadFile(privateKeyPath)
	if err != nil {
		return input.FLB_ERROR
	}
	hostKeyPath := replaceHomeDir(input.FLBPluginConfigKey(plugin, "host_key"))
	if hostKeyPath != "" {
		hostKey, err = os.ReadFile(hostKeyPath)
		if err != nil {
			return input.FLB_ERROR
		}
	}
	logType = input.FLBPluginConfigKey(plugin, "log_type")
	if p := input.FLBPluginConfigKey(plugin, "send_all"); strings.ToLower(p) != "true" {
		lastTime = time.Now().UnixNano()
	}
	log.Printf("[in_twsnmp] twsnmp= '%s'", twsnmp)
	log.Printf("[in_twsnmp] privateKeyPath= '%s'", privateKeyPath)
	log.Printf("[in_twsnmp] hostKeyPath= '%s'", hostKeyPath)
	log.Printf("[in_twsnmp] logType= '%s'", logType)
	log.Printf("[in_twsnmp] lastTime = '%s'", time.Unix(0, lastTime).Format(time.RFC3339))
	if err := getTWSNMPLogs(); err != nil {
		log.Printf("[in_twsnmp] getTWSNMPLogs err:%v", err)
		return input.FLB_ERROR
	}
	return input.FLB_OK
}

func replaceHomeDir(p string) string {
	if !strings.HasPrefix(p, "~") {
		return p
	}
	usr, _ := user.Current()
	return strings.Replace(p, "~", usr.HomeDir, 1)
}

//export FLBPluginInputCallback
func FLBPluginInputCallback(data *unsafe.Pointer, size *C.size_t) int {
	if len(logList) < 1 {
		if err := getTWSNMPLogs(); err != nil || len(logList) < 1 {
			time.Sleep(time.Second * 5)
			return input.FLB_RETRY
		}
	}
	l := logList[0]
	logList = logList[1:]
	flb_time := input.FLBTime{time.Unix(0, l.TimeStamp)}

	entry := []interface{}{flb_time, l.Data}
	enc := input.NewEncoder()
	packed, err := enc.Encode(entry)
	if err != nil {
		log.Printf("[in_twsnmp] encode err:%v", err)
		return input.FLB_ERROR
	}

	length := len(packed)
	*data = C.CBytes(packed)
	*size = C.size_t(length)
	return input.FLB_OK
}

//export FLBPluginInputCleanupCallback
func FLBPluginInputCleanupCallback(data unsafe.Pointer) int {
	return input.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	log.Printf("[in_twsnmp] Exit called")
	return input.FLB_OK
}

func getTWSNMPLogs() error {
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return err
	}
	sshConfig := &ssh.ClientConfig{
		User:    "twsnmp",
		Auth:    []ssh.AuthMethod{},
		Timeout: time.Second * 1,
	}
	sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
	if len(hostKey) > 0 {
		pubkey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(hostKey))
		if err != nil {
			log.Printf("[in_twsnmp] parse host key err:%v", err)
			return err
		}
		sshConfig.HostKeyCallback = ssh.FixedHostKey(pubkey)
	} else {
		sshConfig.HostKeyCallback =
			func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				newKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
				if lastHostKey == "" {
					lastHostKey = newKey
				}
				if lastHostKey != newKey {
					return fmt.Errorf("host key changed")
				}
				return nil
			}
	}
	client, err := ssh.Dial("tcp", twsnmp, sshConfig)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	out, err := session.Output(fmt.Sprintf("get %s %d 1000", logType, lastTime))
	if err != nil {
		return err
	}
	for _, l := range strings.Split(string(out), "\n") {
		l = strings.TrimSpace(l)
		a := strings.SplitN(l, "\t", 2)
		if len(a) != 2 {
			continue
		}
		t, err := strconv.ParseInt(a[0], 10, 64)
		if err != nil {
			continue
		}
		d := new(map[string]interface{})
		err = json.Unmarshal([]byte(a[1]), d)
		if err != nil {
			continue
		}
		e := &logEnt{
			TimeStamp: t,
			Data:      d,
		}
		logList = append(logList, e)
		lastTime = t + 1
	}
	return nil
}

func main() {
}
