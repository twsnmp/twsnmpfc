// パッケージclientは、TWSNMP FCにアクセスするためにWeb APIを利用するライブラリです。
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
	URL                string
	Token              string
	InsecureSkipVerify bool
	Timeout            int
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
