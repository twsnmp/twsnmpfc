package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rakyll/statik/fs"
	"github.com/twsnmp/twsnmpfc/logger"
	"github.com/twsnmp/twsnmpfc/notify"
	_ "github.com/twsnmp/twsnmpfc/statik"

	"github.com/twsnmp/twsnmpfc/backend"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/ping"
	"github.com/twsnmp/twsnmpfc/polling"
	"github.com/twsnmp/twsnmpfc/report"
	"github.com/twsnmp/twsnmpfc/webapi"
)

// From ENV or Command Args
var dataStorePath string
var password string
var port string
var host string
var ip string
var tls bool
var local bool
var cpuprofile string
var memprofile string
var restore string
var pingMode string
var timeout int
var compact string
var backupPath string
var copyBackup bool
var version = "vx.x.x"
var commit = ""

var trapPort = 162
var netflowPort = 2055
var syslogPort = 514
var sflowPort = 6343

var resetPassword bool

var saveMapInterval = -1

func init() {
	flag.StringVar(&dataStorePath, "datastore", "./datastore", "Path to Data Store directory")
	flag.StringVar(&password, "password", "twsnmpfc!", "Master Password")
	flag.StringVar(&port, "port", "8080", "port")
	flag.StringVar(&host, "host", "", "Host Name for TLS Cert")
	flag.StringVar(&restore, "restore", "", "Restore DB file name")
	flag.StringVar(&ip, "ip", "", "IP Address for TLS Cert")
	flag.StringVar(&pingMode, "ping", "", "ping mode icmp or udp")
	flag.BoolVar(&tls, "tls", false, "Use TLS")
	flag.BoolVar(&local, "local", false, "Local only")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&memprofile, "memprofile", "", "write memory profile to `file`")
	flag.IntVar(&timeout, "timeout", 24, "session timeout 0 is unlimit")
	flag.StringVar(&backupPath, "backup", "", "Backup path")
	flag.StringVar(&compact, "compact", "", "DB Conmact path")
	flag.BoolVar(&copyBackup, "copybackup", false, "Use copy mode on backup")
	flag.IntVar(&trapPort, "trapPort", 162, "snmp trap port")
	flag.IntVar(&netflowPort, "netflowPort", 2055, "netflow port")
	flag.IntVar(&syslogPort, "syslogPort", 514, "syslog port")
	flag.IntVar(&sflowPort, "sflowPort", 6343, "sflow port")
	flag.IntVar(&saveMapInterval, "saveMap", -1, "Save Map Interval default: windows=5min,other=60min")
	flag.BoolVar(&resetPassword, "resetPassword", false, "Reset user:password to twsnmp:twsnmp")
	flag.VisitAll(func(f *flag.Flag) {
		if s := os.Getenv("TWSNMPFC_" + strings.ToUpper(f.Name)); s != "" {
			f.Value.Set(s)
		}
	})
	flag.Parse()
}

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().Format("2006-01-02T15:04:05.999 ") + string(bytes))
}

func main() {
	st := time.Now()
	log.Printf("start twsnmpfc version=%s(%s)", version, commit)
	log.Println("config")
	flag.VisitAll(func(f *flag.Flag) {
		// password以外の起動パラメータ、環境変数をログに記録
		if f.Name != "password" {
			log.Printf("%s='%s'", f.Name, f.Value.String())
		}
	})
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatalf("create CPU profile err=%v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("start CPU profile err=%v", err)
		}
		defer pprof.StopCPUProfile()
	}
	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Fatalf("create memory profile  err=%v", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatalf("write memory profile err=%v", err)
		}
	}
	datastore.BackupPath = backupPath
	datastore.CopyBackup = copyBackup
	if restore != "" {
		if err := datastore.RestoreDB(dataStorePath, restore); err != nil {
			log.Fatalf("restore db err=%v", err)
		} else {
			log.Println("restore db done")
		}
		os.Exit(0)
	}
	if compact != "" {
		st := time.Now()
		if err := datastore.CompactDB(dataStorePath, compact); err != nil {
			log.Fatalf("compact db err=%v", err)
		} else {
			log.Printf("compact db done dur=%v", time.Since(st))
		}
		os.Exit(0)
	}
	if resetPassword {
		if err := datastore.ResetPassword(dataStorePath); err != nil {
			log.Fatalf("reset password err=%v", err)
		}
		log.Println("reset password")
		os.Exit(0)
	}
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
	statikFS, err := fs.New()
	if err != nil {
		log.Fatalf("no statik fs err=%v", err)
	}
	log.Println("call datastore.Init")
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	if err = datastore.Init(ctx, dataStorePath, statikFS, wg); err != nil {
		log.Fatalf("init db err=%v", err)
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: "TWSNMP FC起動",
	})
	log.Println("call ping.Start")
	if err = ping.Start(ctx, wg, pingMode); err != nil {
		log.Fatalf("start ping err=%v", err)
	}
	log.Println("call report.Start")
	if err = report.Start(ctx, wg); err != nil {
		log.Fatalf("start report err=%v", err)
	}
	log.Println("call logger.Start")
	if err = logger.Start(ctx, wg, trapPort, netflowPort, syslogPort, sflowPort); err != nil {
		log.Fatalf("start logger err=%v", err)
	}
	log.Println("call polling.Start")
	if err = polling.Start(ctx, wg); err != nil {
		log.Fatalf("start polling err=%v", err)
	}
	log.Println("call backend.Start")
	if saveMapInterval < 0 {
		if runtime.GOOS == "windows" {
			saveMapInterval = 5
		} else {
			saveMapInterval = 60 * 6
		}
	}
	backend.SaveMapInterval = saveMapInterval
	log.Printf("set SaveMapInterval=%d", saveMapInterval)
	if err = backend.Start(ctx, dataStorePath, version, wg); err != nil {
		log.Fatalf("start backend err=%v", err)
	}
	log.Println("call notify.Start")
	if err = notify.Start(ctx, wg); err != nil {
		log.Fatalf("start notify err=%v", err)
	}
	quit := make(chan os.Signal, 1)
	log.Println("call webapi.Start")
	w := &webapi.WebAPI{
		Statik:        http.FileServer(statikFS),
		Port:          port,
		UseTLS:        tls,
		Local:         local,
		IP:            ip,
		Host:          host,
		Password:      password,
		Version:       fmt.Sprintf("%s(%s)", version, commit),
		DataStorePath: dataStorePath,
		QuitSignal:    quit,
		Timeout:       timeout,
	}
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go webapi.Start(w)
	if local {
		time.Sleep(3 * time.Second)
		openURL(fmt.Sprintf("http://127.0.0.1:%s", port))
	}
	if runtime.GOOS == "windows" {
		go stopper(quit)
	}
	sig := <-quit
	stop := time.Now()
	log.Printf("signal twsnmpfc signal=%v", sig)
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: "TWSNMP FC停止",
	})
	webapi.Stop()
	cancel()
	wg.Wait()
	datastore.CloseDB()
	log.Printf("stop twsnmpfc dur=%v stop=%v", time.Since(st), time.Since(stop))
}

func openURL(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Printf("open url=%s err=%v", url, err)
		return err
	}
	return err
}

func stopper(sig chan os.Signal) {
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8080,
	}
	var l *net.UDPConn
	var err error
	for i := 8080; i < 8180; i++ {
		udpAddr.Port = i
		l, err = net.ListenUDP("udp", udpAddr)
		if err == nil {
			log.Printf("stopper port=%d", i)
			break
		}
		log.Println(err)
	}
	if l == nil {
		log.Println(err)
		return
	}
	defer l.Close()
	for {
		b := make([]byte, 256)
		n, _, err := l.ReadFrom(b)
		if err == nil && n > 0 {
			rpid := string(b[:n])
			fmt.Printf("%d '%s' '%d'", n, rpid, os.Getpid())
			pid := fmt.Sprintf("%d", os.Getpid())
			if pid == rpid {
				sig <- os.Interrupt
			}
		}
	}
}
