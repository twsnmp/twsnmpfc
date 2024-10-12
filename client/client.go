// Package clinet : Web APIを利用したクライントライブラリです。
package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TWSNMPApiは、TWSNMP FCと通信するためのデータ構造です。
type TWSNMPApi struct {
	URL                string // TWSNMP FCのURL
	Token              string // JWTアクセストークン
	InsecureSkipVerify bool   // HTTPSで通信する時にサーバー証明書の検証を行わない
	Timeout            int    // タイムアウト
}

type selectEntWebAPI struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

// NewClientは新しいクライアントを作成します。
func NewClient(url string) *TWSNMPApi {
	return &TWSNMPApi{
		URL: url,
	}
}

func (a *TWSNMPApi) twsnmpHTTPClient() *http.Client {
	if a.Timeout < 1 {
		a.Timeout = 30
	}
	return &http.Client{
		Timeout: time.Duration(a.Timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: a.InsecureSkipVerify,
			},
		},
	}
}

type loginParam struct {
	UserID   string
	Password string
}

// LoginはTWSNMP FCにログインします。
func (a *TWSNMPApi) Login(user, password string) error {
	lp := &loginParam{
		UserID:   user,
		Password: password,
	}
	j, err := json.Marshal(lp)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(
		"POST",
		a.URL+"/login",
		bytes.NewBuffer([]byte(j)),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := a.twsnmpHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	lr := make(map[string]string)
	if err := json.Unmarshal(body, &lr); err != nil {
		return err
	}
	a.Token = lr["token"]
	return nil
}

// GetはTWSNMP FCにGETリクエストを送信します。
func (a *TWSNMPApi) Get(path string) ([]byte, error) {
	req, err := http.NewRequest(
		"GET",
		a.URL+path,
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.Token)
	client := a.twsnmpHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 200 {
		return nil, fmt.Errorf("resp code=%d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// PostはTWSNMP FCにPOSTリクエストを送信します。
func (a *TWSNMPApi) Post(path string, data []byte) ([]byte, error) {
	req, err := http.NewRequest(
		"POST",
		a.URL+path,
		bytes.NewBuffer([]byte(data)),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.Token)
	client := a.twsnmpHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 200 {
		return nil, fmt.Errorf("resp code=%d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// DeleteはTWSNMP FCにDELETEリクエストを送信します。
func (a *TWSNMPApi) Delete(path string) error {
	req, err := http.NewRequest(
		"DELETE",
		a.URL+path,
		nil,
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.Token)
	client := a.twsnmpHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 204 {
		return fmt.Errorf("twsnmp api delete code=%d", resp.StatusCode)
	}
	return nil
}

// From datastore

type NodeEnt struct {
	ID           string
	Name         string
	Descr        string
	Icon         string
	Image        string
	State        string
	X            int
	Y            int
	IP           string
	IPv6         string
	MAC          string
	SnmpMode     string
	Community    string
	User         string
	Password     string
	GNMIUser     string
	GNMIPassword string
	GNMIEncoding string
	GNMIPort     string
	PublicKey    string
	URL          string
	AddrMode     string
	AutoAck      bool
}

type EventLogEnt struct {
	Time     int64 // UnixNano()
	Type     string
	Level    string
	NodeName string
	NodeID   string
	Event    string
}

type LogEnt struct {
	Time int64 // UnixNano()
	Type string
	Log  string
}

type LogFilterEnt struct {
	StartTime string
	EndTime   string
	Filter    string
	LogType   string
}

type SFlowCounterEnt struct {
	Remote string
	Type   string
	Data   string
}

type PollingEnt struct {
	ID           string
	Name         string
	NodeID       string
	Type         string
	Mode         string
	Params       string
	Filter       string
	Extractor    string
	Script       string
	Level        string
	PollInt      int
	Timeout      int
	Retry        int
	LogMode      int
	NextTime     int64
	LastTime     int64
	Result       map[string]interface{}
	State        string
	FailAction   string
	RepairAction string
}

type PollingLogEnt struct {
	Time      int64 // UnixNano()
	PollingID string
	State     string
	Result    map[string]interface{}
}

type DeviceEnt struct {
	ID         string // MAC Addr
	Name       string
	IP         string
	NodeID     string
	Vendor     string
	Score      float64
	ValidScore bool
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

type UserClientEnt struct {
	Total int32
	Ok    int32
}

type UserEnt struct {
	ID           string // User ID + Server
	UserID       string
	Server       string
	ServerName   string
	ServerNodeID string
	ClientMap    map[string]UserClientEnt
	Total        int
	Ok           int
	Score        float64
	ValidScore   bool
	Penalty      int64
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

type ServerEnt struct {
	ID           string //  ID Server
	Server       string
	Services     map[string]int64
	Count        int64
	Bytes        int64
	ServerName   string
	ServerNodeID string
	Loc          string
	Score        float64
	ValidScore   bool
	Penalty      int64
	TLSInfo      string
	NTPInfo      string
	DHCPInfo     string
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

type FlowEnt struct {
	ID           string // ID Client:Server
	Client       string
	Server       string
	Services     map[string]int64
	Count        int64
	Bytes        int64
	ClientName   string
	ClientNodeID string
	ClientLoc    string
	ServerName   string
	ServerNodeID string
	ServerLoc    string
	Score        float64
	ValidScore   bool
	Penalty      int64
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

type IPReportEnt struct {
	IP         string
	MAC        string
	Name       string
	NodeID     string
	Loc        string
	Vendor     string
	Count      int64
	Change     int64
	Score      float64
	ValidScore bool
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

type SensorEnt struct {
	ID        string // Host + Type + Param
	Host      string
	Type      string // twpcap,twwinlog....
	Param     string
	Total     int64
	Send      int64
	State     string
	Ignore    bool
	Stats     []SensorStatsEnt
	Monitors  []SensorMonitorEnt
	FirstTime int64
	LastTime  int64
}

type SensorStatsEnt struct {
	Time     int64
	Total    int64
	Count    int64
	PS       float64
	Send     int64
	LastSend int64
}

type SensorMonitorEnt struct {
	Time    int64
	CPU     float64
	Mem     float64
	Load    float64
	Process int64
	Recv    int64
	Sent    int64
	TxSpeed float64
	RxSpeed float64
}

type MonitorDataEnt struct {
	CPU   float64
	Mem   float64
	Disk  float64
	Load  float64
	Bytes float64
	Net   float64
	Proc  int
	Conn  int
	At    int64
}
