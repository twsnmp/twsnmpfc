package client

import (
	"encoding/json"
	"fmt"
)

// GetDevices はTWSNMP FCからLANデバイスレポートを取得します。
func (a *TWSNMPApi) GetDevices() ([]*DeviceEnt, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	data, err := a.Get("/api/report/devices")
	if err != nil {
		return nil, err
	}
	devices := []*DeviceEnt{}
	err = json.Unmarshal(data, &devices)
	return devices, err
}

// GetUsers はTWSNMP FCからユーザーレポートを取得します。
func (a *TWSNMPApi) GetUsers() ([]*UserEnt, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	data, err := a.Get("/api/report/users")
	if err != nil {
		return nil, err
	}
	users := []*UserEnt{}
	err = json.Unmarshal(data, &users)
	return users, err
}

// GetServers はTWSNMP FCからサーバーレポートを取得します。
func (a *TWSNMPApi) GetServers() ([]*ServerEnt, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	data, err := a.Get("/api/report/servers")
	if err != nil {
		return nil, err
	}
	servers := []*ServerEnt{}
	err = json.Unmarshal(data, &servers)
	return servers, err
}

// GetFlows はTWSNMP FCからフローレポートを取得します。
func (a *TWSNMPApi) GetFlows() ([]*FlowEnt, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	data, err := a.Get("/api/report/flows")
	if err != nil {
		return nil, err
	}
	flows := []*FlowEnt{}
	err = json.Unmarshal(data, &flows)
	return flows, err
}

// GetIPs はTWSNMP FCからIPレポートを取得します。
func (a *TWSNMPApi) GetIPs() ([]*IPReportEnt, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	data, err := a.Get("/api/report/flows")
	if err != nil {
		return nil, err
	}
	ips := []*IPReportEnt{}
	err = json.Unmarshal(data, &ips)
	return ips, err
}

// GetSensors はTWSNMP FCからセンサーレポートを取得します。
func (a *TWSNMPApi) GetSensors() ([]*SensorEnt, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	data, err := a.Get("/api/report/sensors")
	if err != nil {
		return nil, err
	}
	sensors := []*SensorEnt{}
	err = json.Unmarshal(data, &sensors)
	return sensors, err
}

// GetMonitor はTWSNMP FCからモニターレポートを取得します。
func (a *TWSNMPApi) GetMonitor() ([]*MonitorDataEnt, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	data, err := a.Get("/api/monitor")
	if err != nil {
		return nil, err
	}
	monitor := []*MonitorDataEnt{}
	err = json.Unmarshal(data, &monitor)
	return monitor, err
}

// AiListEntWebAPI は、AI分析リストのデータ構造です。
type AiListEntWebAPI struct {
	ID          string  // 内部ID
	NodeID      string  // ノードの内部ID
	NodeName    string  // ノード名
	PollingName string  // ポーリング名
	Score       float64 // スコア
	Count       int     // 件数
	LastTime    int64   // 最終分析日時
}

// GetAIList はTWSNMP FCからAI分析リストを取得します。
func (a *TWSNMPApi) GetAIList() ([]*AiListEntWebAPI, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	data, err := a.Get("/api/report/ailist")
	if err != nil {
		return nil, err
	}
	aiList := []*AiListEntWebAPI{}
	err = json.Unmarshal(data, &aiList)
	return aiList, err
}

// DeleteReport は,指定された種別とIDのレポートを削除します。
func (a *TWSNMPApi) DeleteReport(t, id string) error {
	if a.Token == "" {
		return fmt.Errorf("not login")
	}
	return a.Delete("/api/report/" + t + "/" + id)
}

// ResetReport は,指定された種別とIDのレポートのスコアを再計算します。
func (a *TWSNMPApi) ResetReport(t string) error {
	if a.Token == "" {
		return fmt.Errorf("not login")
	}
	_, err := a.Post("/api/report/"+t+"/reset", []byte{})
	return err
}
