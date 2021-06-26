package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"
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

var version = "v1.2.0"
var commit = ""

func init() {
	flag.StringVar(&dataStorePath, "datastore", "./datastore", "Path to Data Store directory")
	flag.StringVar(&password, "password", "twsnmpfc!", "Master Password")
	flag.StringVar(&port, "port", "8080", "port")
	flag.StringVar(&host, "host", "", "Host Name for TLS Cert")
	flag.StringVar(&restore, "restore", "", "Restore DB file name")
	flag.StringVar(&ip, "ip", "", "IP Address for TLS Cert")
	flag.BoolVar(&tls, "tls", false, "Use TLS")
	flag.BoolVar(&local, "local", false, "Local only")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&memprofile, "memprofile", "", "write memory profile to `file`")
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
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatalf("could not create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}
	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Fatalf("could not create memory profile: %v", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatalf("could not write memory profile:%v", err)
		}
	}
	if restore != "" {
		if err := datastore.RestoreDB(dataStorePath, restore); err != nil {
			log.Fatalf("restore err=%v", err)
		} else {
			log.Println("restore done")
		}
		os.Exit(0)
	}
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
	statikFS, err := fs.New()
	if err != nil {
		log.Fatalln(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err = datastore.Init(ctx, dataStorePath, statikFS); err != nil {
		log.Fatalln(err)
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: "TWSNMP FC起動",
	})
	if err = ping.Start(ctx); err != nil {
		log.Fatalln(err)
	}
	if err = report.Start(ctx); err != nil {
		log.Fatalln(err)
	}
	if err = logger.Start(ctx); err != nil {
		log.Fatalln(err)
	}
	if err = polling.Start(ctx); err != nil {
		log.Fatalln(err)
	}
	if err = backend.Start(ctx, dataStorePath, version); err != nil {
		log.Fatalln(err)
	}
	if err = notify.Start(ctx); err != nil {
		log.Fatalln(err)
	}
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
	}
	log.Printf("version=%s", w.Version)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go webapi.Start(w)
	if local {
		time.Sleep(3 * time.Second)
		openURL(fmt.Sprintf("http://127.0.0.1:%s", port))
	}
	<-quit
	log.Println("quit by signal")
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: "TWSNMP FC停止",
	})
	webapi.Stop()
	cancel()
	time.Sleep(time.Second * 2)
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
		log.Printf("openUrl err=%v", err)
		return err
	}
	return err
}
