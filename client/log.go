// パッケージclientは、TWSNMP FCにアクセスするためにWeb APIを利用するライブラリです。
package client

import (
	"encoding/json"
	"fmt"

	"github.com/twsnmp/twsnmpfc/datastore"
)

// EventLogFilterは、イベントログの検索条件です。
type EventLogFilter struct {
	Level     string
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
	Type      string
	NodeID    string
	Event     string
}

// EventLogsWebAPIは、イベントログの検索結果です。
type EventLogsWebAPI struct {
	EventLogs []*datastore.EventLogEnt
	NodeList  []selectEntWebAPI
}

// GetEventLogsはTWSNMP FCからイベントログを取得します。
func (a *TWSNMPApi) GetEventLogs(filter *EventLogFilter) (*EventLogsWebAPI, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	j, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	data, err := a.Post("/api/log/eventlogs", j)
	if err != nil {
		return nil, err
	}
	logs := EventLogsWebAPI{}
	err = json.Unmarshal(data, &logs)
	return &logs, err
}

// SyslogFilterは、syslogの検索条件です。
type SyslogFilter struct {
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
	Level     string
	Type      string
	Host      string
	Tag       string
	Message   string
	Extractor string
	NextTime  int64
	Filter    int
}

// SyslogWebAPIは、syslogの検索結果です。
type SyslogWebAPI struct {
	Logs          []*SyslogWebAPILogEnt
	ExtractHeader []string
	ExtractDatas  [][]string
	NextTime      int64
	Process       int
	Filter        int
	Limit         int
}

// SyslogWebAPILogEntは、syslogのデータ構造です。
type SyslogWebAPILogEnt struct {
	Time     int64
	Level    string
	Host     string
	Type     string
	Tag      string
	Message  string
	Severity int
	Facility int
}

// GetSyslogsはTWSNMP FCからsyslogを取得します。
func (a *TWSNMPApi) GetSyslogs(filter *SyslogFilter) (*SyslogWebAPI, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	j, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	data, err := a.Post("/api/log/syslog", j)
	if err != nil {
		return nil, err
	}
	logs := SyslogWebAPI{}
	err = json.Unmarshal(data, &logs)
	return &logs, err
}

// SnmpTrapFilterは、SNMP TRAPログの検索条件です。
type SnmpTrapFilter struct {
	StartDate   string
	StartTime   string
	EndDate     string
	EndTime     string
	FromAddress string
	TrapType    string
	Variables   string
}

// SnmpTrapWebAPIは、SNMP TRAPログです。
type SnmpTrapWebAPI struct {
	Time        int64
	FromAddress string
	TrapType    string
	Variables   string
}

// GetSnmpTrapsはTWSNMP FCからSNMP Trapログを取得します。
func (a *TWSNMPApi) GetSnmpTraps(filter *SnmpTrapFilter) ([]*SnmpTrapWebAPI, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	j, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	data, err := a.Post("/api/log/snmptrap", j)
	if err != nil {
		return nil, err
	}
	logs := []*SnmpTrapWebAPI{}
	err = json.Unmarshal(data, &logs)
	return logs, err
}

// NetflowFilterは、NetFlow/IPFIXの検索条件です。
type NetflowFilter struct {
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
	SrcDst    bool
	IP        string
	Port      string
	SrcIP     string
	SrcPort   string
	DstIP     string
	DstPort   string
	Protocol  string
	TCPFlag   string
	NextTime  int64
	Filter    int
}

// NetflowWebAPIは、NetFlow/IPFIXログの検索結果です。
type NetflowWebAPI struct {
	Logs     []*NetflowWebAPILogEnt
	NextTime int64
	Process  int
	Filter   int
	Limit    int
}

// NetflowWebAPILogEntは、NetFlow/IPFIXログのデータ構造です。
type NetflowWebAPILogEnt struct {
	Time     int64
	Src      string
	SrcIP    string
	SrcMAC   string
	SrcPort  int
	DstIP    string
	DstMAC   string
	DstPort  int
	Dst      string
	Protocol string
	TCPFlags string
	Packets  int64
	Bytes    int64
	Duration float64
}

// GetNetFlowはTWSNMP FCからNetFlowログを取得します。
func (a *TWSNMPApi) GetNetFlow(filter *NetflowFilter) (*NetflowWebAPI, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	j, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	data, err := a.Post("/api/log/netflow", j)
	if err != nil {
		return nil, err
	}
	r := NetflowWebAPI{}
	err = json.Unmarshal(data, &r)
	return &r, err
}

// GetIPFIXはTWSNMP FCからIPFIXログを取得します。
func (a *TWSNMPApi) GetIPFIX(filter *NetflowFilter) (*NetflowWebAPI, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	j, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	data, err := a.Post("/api/log/ipfix", j)
	if err != nil {
		return nil, err
	}
	r := NetflowWebAPI{}
	err = json.Unmarshal(data, &r)
	return &r, err
}

// SFlowFilterはSFlowログ検索条件です。
type SFlowFilter struct {
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
	SrcDst    bool
	IP        string
	Port      string
	SrcIP     string
	SrcPort   string
	DstIP     string
	DstPort   string
	Protocol  string
	TCPFlag   string
	Reason    string
	NextTime  int64
	Filter    int
}

// SFlowWebAPIは、SFlowログの検索結果です。
type SFlowWebAPI struct {
	Logs     []*SFlowWebAPILogEnt
	NextTime int64
	Process  int
	Filter   int
	Limit    int
}

// SFlowWebAPILogEntは、SFlowログです。
type SFlowWebAPILogEnt struct {
	Time     int64
	Src      string
	SrcIP    string
	SrcMAC   string
	SrcPort  int
	DstIP    string
	DstMAC   string
	DstPort  int
	Dst      string
	Protocol string
	TCPFlags string
	Bytes    int64
	Reason   int
}

// GetSFlowはTWSNMP FCからSFlowログを取得します。
func (a *TWSNMPApi) GetSFlow(filter *SFlowFilter) ([]*SFlowWebAPI, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	j, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	data, err := a.Post("/api/log/sflow", j)
	if err != nil {
		return nil, err
	}
	logs := []*SFlowWebAPI{}
	err = json.Unmarshal(data, &logs)
	return logs, err
}

// SFlowCounterFilterは、SFlow Counterログの検索条件です。
type SFlowCounterFilter struct {
	Remote    string
	Type      string
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
	NextTime  int64
	Filter    int
}

// SFlowCounterWebAPIは、SFlow Counterログの検索結果です。
type SFlowCounterWebAPI struct {
	Logs     []*SFlowCounterWebAPILogEnt
	NextTime int64
	Process  int
	Filter   int
	Limit    int
}

// SFlowCounterWebAPILogEntは、SFlowCounterログです。
type SFlowCounterWebAPILogEnt struct {
	Time int64
	datastore.SFlowCounterEnt
}

// GetSFlowCounterはTWSNMP FCからSFlow Counterログを取得します。
func (a *TWSNMPApi) GetSFlowCounter(filter *SFlowCounterFilter) (*SFlowCounterWebAPI, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	j, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	data, err := a.Post("/api/log/sflowCounter", j)
	if err != nil {
		return nil, err
	}
	logs := SFlowCounterWebAPI{}
	err = json.Unmarshal(data, &logs)
	return &logs, err
}

// ArpFilterは、ARPログの検索条件を指定します。
type ArpFilter struct {
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
	IP        string
	MAC       string
}

// ArpWebAPIは、ARPログです。
type ArpWebAPI struct {
	Time      int64
	State     string
	IP        string
	MAC       string
	Vendor    string
	OldMAC    string
	OldVendor string
}

// GetArpLogsはTWSNMP FCからARPログを取得します。
func (a *TWSNMPApi) GetArpLogs(filter *ArpFilter) ([]*ArpWebAPI, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	j, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	data, err := a.Post("/api/log/arp", j)
	if err != nil {
		return nil, err
	}
	logs := []*ArpWebAPI{}
	err = json.Unmarshal(data, &logs)
	return logs, err
}

// TimeFilterは、ポーリングログの時間範囲の条件を指定します。
type TimeFilter struct {
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
}

// GetPollingLogsはTWSNMP FCからポーリングログを取得します。
func (a *TWSNMPApi) GetPollingLogs(id string, filter *TimeFilter) ([]*datastore.PollingLogEnt, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	j, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	data, err := a.Post("/api/pollingLogs/"+id, j)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%s", string(data))
	logs := []*datastore.PollingLogEnt{}
	err = json.Unmarshal(data, &logs)
	return logs, err
}

// DeleteLogは、idで指定されたログを全て削除します。
func (a *TWSNMPApi) DeleteLog(id string) error {
	if a.Token == "" {
		return fmt.Errorf("not login")
	}
	if id == "arp" {
		return a.Delete("/api/arp")
	}
	return a.Delete("/api/log/" + id)
}
