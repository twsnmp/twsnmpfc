// パッケージclientは、TWSNMP FCにアクセスするためにWeb APIを利用するライブラリです。
package client

import (
	"encoding/json"
	"fmt"

	"github.com/twsnmp/twsnmpfc/datastore"
)

// GetNodesはTWSNMP FCからノードリストを取得します。
func (a *TWSNMPApi) GetNodes() ([]*datastore.NodeEnt, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	data, err := a.Get("/api/nodes")
	if err != nil {
		return nil, err
	}
	nodes := []*datastore.NodeEnt{}
	err = json.Unmarshal(data, &nodes)
	return nodes, err
}

// UpdateNodeは、ノードを追加または削除します。
func (a *TWSNMPApi) UpdateNode(node *datastore.NodeEnt) error {
	if a.Token == "" {
		return fmt.Errorf("not login")
	}
	j, err := json.Marshal(node)
	if err != nil {
		return err
	}
	_, err = a.Post("/api/node/update", j)
	return err
}

// DeleteNodesは、ノードを削除します。
func (a *TWSNMPApi) DeleteNodes(ids []string) error {
	if a.Token == "" {
		return fmt.Errorf("not login")
	}
	j, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	_, err = a.Post("/api/nodes/delete", j)
	return err
}

type PollingsWebAPI struct {
	Pollings []*datastore.PollingEnt
	NodeList []selectEntWebAPI
}

// GetPollingsはTWSNMP FCからポーリングリストを取得します。
func (a *TWSNMPApi) GetPollings() (*PollingsWebAPI, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not login")
	}
	data, err := a.Get("/api/pollings")
	if err != nil {
		return nil, err
	}
	pollings := PollingsWebAPI{}
	err = json.Unmarshal(data, &pollings)
	return &pollings, err
}

// UpdatePollingは、ポーリングの追加または更新します。
func (a *TWSNMPApi) UpdatePolling(polling *datastore.PollingEnt) error {
	if a.Token == "" {
		return fmt.Errorf("not login")
	}
	j, err := json.Marshal(polling)
	if err != nil {
		return err
	}
	if polling.ID == "" {
		_, err = a.Post("/api/polling/add", j)
	} else {
		_, err = a.Post("/api/polling/update", j)
	}
	return err
}

// DeletePollingsは、ポーリングを削除します。
func (a *TWSNMPApi) DeletePollings(ids []string) error {
	if a.Token == "" {
		return fmt.Errorf("not login")
	}
	j, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	_, err = a.Post("/api/pollings/delete", j)
	return err
}
