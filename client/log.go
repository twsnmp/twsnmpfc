package client

import (
	"encoding/json"
	"fmt"

	"github.com/twsnmp/twsnmpfc/datastore"
)

// EventLogFilterは、イベントログの検索条件です。
type EventLogFilter struct {
	Level     string // ログのレベル
	StartDate string // 検索の開始日 例:2006-01-02
	StartTime string // 検索の開始時刻 例: 15:04
	EndDate   string // 検索の終了日 例:2006-01-02
	EndTime   string // 検索の終了時刻 例: 15:04
	Type      string // ログのタイプ
	NodeID    string // ログの対象ノードのID
	Event     string // イベントに含まれる文字列（正規表現)
}

// EventLogsWebAPIは、イベントログの検索結果です。
type EventLogsWebAPI struct {
	EventLogs []*datastore.EventLogEnt // イベントログの検索結果
	NodeList  []selectEntWebAPI        // ノード名選択肢
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
	StartDate string // 検索の開始日 例:2006-01-02
	StartTime string // 検索の開始時刻 例: 15:04
	EndDate   string // 検索の終了日 例:2006-01-02
	EndTime   string // 検索の終了時刻 例: 15:04
	Level     string // レベル
	Type      string // ログのタイプ
	Host      string // ログの送信元ホスト
	Tag       string // ログのタグ
	Message   string // ログのメッセージ
	Extractor string // 抽出パターン
	NextTime  int64  // 継続ログの開始時刻(nano Sec)
	Filter    int    // ログの累積件数
}

// SyslogWebAPIは、syslogの検索結果です。
type SyslogWebAPI struct {
	Logs          []*SyslogWebAPILogEnt // sysylogの検索結果
	ExtractHeader []string              // 抽出データのヘッダ
	ExtractDatas  [][]string            // 抽出データ
	NextTime      int64                 // 継続検索の開始時刻(nano Sec)
	Process       int                   // ログの総数
	Filter        int                   // フィルター後の件数
	Limit         int                   // 検索上限
}

// SyslogWebAPILogEntは、syslogのデータ構造です。
type SyslogWebAPILogEnt struct {
	Time     int64  // タイムスタンプ(nano Sec)
	Level    string // レベル
	Host     string // 送信元ホスト
	Type     string // ログの種別
	Tag      string // タグ
	Message  string // メッセージ
	Severity int    // Severity
	Facility int    // Facility
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
	StartDate   string // 検索の開始日 例:2006-01-02
	StartTime   string // 検索の開始時刻 例: 15:04
	EndDate     string // 検索の終了日 例:2006-01-02
	EndTime     string // 検索の終了時刻 例: 15:04
	FromAddress string // TRAPの送信元アドレス
	TrapType    string // TRAPの種別
	Variables   string // TRAPに付帯したMIB
}

// SnmpTrapWebAPIは、SNMP TRAPログです。
type SnmpTrapWebAPI struct {
	Time        int64  // TRAPの受信日時
	FromAddress string // TRAPの送信元アドレス
	TrapType    string // TRAPの種別
	Variables   string // TRAPに付帯したMIB
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
	StartDate string // 検索の開始日 例:2006-01-02
	StartTime string // 検索の開始時刻 例: 15:04
	EndDate   string // 検索の終了日 例:2006-01-02
	EndTime   string // 検索の終了時刻 例: 15:04	StartDate string
	SrcDst    bool   // 双方向検索
	IP        string // 双方向検索の場合のIPアドレス
	Port      string // 双方向検索の場合のポート番号
	SrcIP     string // 送信元IP
	SrcPort   string // 送信元ポート
	DstIP     string // 宛先IP
	DstPort   string // 宛先ポート
	Protocol  string // プロトコル
	TCPFlag   string // TCPフラグ
	NextTime  int64  // 継続検索の場合の時刻(nano Sec)
	Filter    int    // 累計の検索数
}

// NetflowWebAPIは、NetFlow/IPFIXログの検索結果です。
type NetflowWebAPI struct {
	Logs     []*NetflowWebAPILogEnt // ログの検索結果
	NextTime int64                  // 継続検索の次回の時刻(nano Sec)
	Process  int                    // 総ログ件数
	Filter   int                    // フィルターした件数
	Limit    int                    // 上限
}

// NetflowWebAPILogEntは、NetFlow/IPFIXログのデータ構造です。
type NetflowWebAPILogEnt struct {
	Time     int64   // ログの受信日時
	Src      string  // 送信元ホスト名
	SrcIP    string  // 送信元IP
	SrcMAC   string  // 送信元MACアドレス
	SrcPort  int     // 送信元ポート番号
	DstIP    string  // 宛先IP
	DstMAC   string  // 宛先MACアドレス
	DstPort  int     // 宛先ポート番号
	Dst      string  // 宛先ホスト名
	Protocol string  // プロコトル
	TCPFlags string  // TCPフラグ
	Packets  int64   // パケット数
	Bytes    int64   // バイト数
	Duration float64 // 通信期間
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
	StartDate string // 検索の開始日 例:2006-01-02
	StartTime string // 検索の開始時刻 例: 15:04
	EndDate   string // 検索の終了日 例:2006-01-02
	EndTime   string // 検索の終了時刻 例: 15:04	StartDate string
	SrcDst    bool   // 双方向検索
	IP        string // 双方向検索のIP
	Port      string // 双方向検索のポート
	SrcIP     string // 送信元IP
	SrcPort   string // 送信元ポート番号
	DstIP     string // 宛先IP
	DstPort   string // 宛先ポート
	Protocol  string // プロトコル
	TCPFlag   string // TCPフラグ
	Reason    string // 切断理由
	NextTime  int64  // 継続検索の開始時刻(nano Sec)
	Filter    int    // 累計検索件数
}

// SFlowWebAPIは、SFlowログの検索結果です。
type SFlowWebAPI struct {
	Logs     []*SFlowWebAPILogEnt // 検索結果のログ
	NextTime int64                // 継続検索の次回時刻(nano Sec)
	Process  int                  // 総ログ数
	Filter   int                  // フィルターした件数
	Limit    int                  // 上限
}

// SFlowWebAPILogEntは、SFlowログです。
type SFlowWebAPILogEnt struct {
	Time     int64  // ログ受信日時
	Src      string // 送信元ホスト
	SrcIP    string // 送信元IP
	SrcMAC   string // 送信元MACアドレス
	SrcPort  int    // 送信元ポート番号
	DstIP    string // 宛先IP
	DstMAC   string // 宛先MACアドレス
	DstPort  int    // 宛先ポート番号
	Dst      string // 宛先ホスト名
	Protocol string // プロトコル
	TCPFlags string // TCPフラグ
	Bytes    int64  // バイト数
	Reason   int    // 切断理由
}

// GetSFlowはTWSNMP FCからSFlowログを取得します。
func (a *TWSNMPApi) GetSFlow(filter *SFlowFilter) (*SFlowWebAPI, error) {
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
	logs := SFlowWebAPI{}
	err = json.Unmarshal(data, &logs)
	return &logs, err
}

// SFlowCounterFilterは、SFlow Counterログの検索条件です。
type SFlowCounterFilter struct {
	Remote    string // 送信元
	Type      string // 種別
	StartDate string // 検索の開始日 例:2006-01-02
	StartTime string // 検索の開始時刻 例: 15:04
	EndDate   string // 検索の終了日 例:2006-01-02
	EndTime   string // 検索の終了時刻 例: 15:04
	NextTime  int64  // 継続検索の開始時刻(nano Sec)
	Filter    int    // 累計検索件数
}

// SFlowCounterWebAPIは、SFlow Counterログの検索結果です。
type SFlowCounterWebAPI struct {
	Logs     []*SFlowCounterWebAPILogEnt // 検索結果のログ
	NextTime int64                       // 継続検索の次回時刻
	Process  int                         // 総ログ数
	Filter   int                         // フィルターした件数
	Limit    int                         // 上限
}

// SFlowCounterWebAPILogEntは、SFlowCounterログです。
type SFlowCounterWebAPILogEnt struct {
	Time                      int64 // 受信時刻
	datastore.SFlowCounterEnt       // カウンターサンプルのデータ
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
	StartDate string // 検索の開始日 例:2006-01-02
	StartTime string // 検索の開始時刻 例: 15:04
	EndDate   string // 検索の終了日 例:2006-01-02
	EndTime   string // 検索の終了時刻 例: 15:04
	IP        string // IPアドレス
	MAC       string // MACアドレス
}

// ArpWebAPIは、ARPログです。
type ArpWebAPI struct {
	Time      int64  // 検知日時
	State     string // 状態
	IP        string // IPアドレス
	MAC       string // MACアドレス
	Vendor    string // ベンダー
	OldMAC    string // 前のMACアドレス
	OldVendor string // 前のベンダー
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
	StartDate string // 検索の開始日 例:2006-01-02
	StartTime string // 検索の開始時刻 例: 15:04
	EndDate   string // 検索の終了日 例:2006-01-02
	EndTime   string // 検索の終了時刻 例: 15:04
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
