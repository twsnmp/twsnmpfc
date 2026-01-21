package polling

import (
	"fmt"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/knadh/go-pop3"
)

type emailStatsEnt struct {
	Count     int
	Size      int64
	LastDate  string
	LastMsgID string
}

func doPollingEMail(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		setPollingError("email", pe, fmt.Errorf("node not found"))
		return
	}
	var stats *emailStatsEnt
	var err error
	st := time.Now().UnixNano()
	if strings.HasPrefix(pe.Params, "imap") {
		stats, err = getIMAPStats(pe)
	} else if strings.HasPrefix(pe.Params, "pop3") {
		stats, err = getPOP3Stats(pe)
	} else {
		setPollingError("email", pe, fmt.Errorf("email protocol not found"))
		return
	}
	if pe.Result == nil {
		pe.Result = make(map[string]interface{})
	}
	if pe.Mode == "login" {
		if err == nil {
			pe.Result["rtt"] = float64(time.Now().UnixNano() - st)
			delete(pe.Result, "error")
			setPollingState(pe, "normal")
			return
		}
		pe.Result["error"] = err.Error()
		setPollingState(pe, pe.Level)
		return
	}
	if err != nil {
		setPollingError("email", pe, err)
		return
	}
	delete(pe.Result, "error")
	pe.Result["newMsg"] = stats.LastMsgID != pe.Result["lastMsgID"]
	pe.Result["lastMsgID"] = stats.LastMsgID
	pe.Result["lastDate"] = stats.LastDate
	pe.Result["count"] = float64(stats.Count)
	pe.Result["size"] = float64(stats.Size)
	if pe.Script == "" {
		setPollingState(pe, "normal")
		return
	}
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	value, err := vm.Run(pe.Script)
	if err == nil {
		if ok, _ := value.ToBoolean(); !ok {
			setPollingState(pe, pe.Level)
			return
		}
		setPollingState(pe, "normal")
		return
	}
	setPollingError("email", pe, err)
}

// IMAPプロトコルによるポーリング処理
func getIMAPStats(pe *datastore.PollingEnt) (*emailStatsEnt, error) {
	// URLのパース
	u, err := url.Parse(pe.Params)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	server := u.Host
	// ポート番号が指定されていない場合のデフォルト設定
	if !strings.Contains(server, ":") {
		if u.Scheme == "imaps" {
			server += ":993"
		} else {
			server += ":143"
		}
	}
	var c *imapclient.Client
	// 接続の確立 (TLSまたはStartTLS)
	if u.Scheme == "imaps" || strings.HasSuffix(server, ":993") {
		c, err = imapclient.DialTLS(server, nil)
	} else {
		c, err = imapclient.DialStartTLS(server, nil)
		if err != nil {
			// StartTLSが失敗した場合は、非セキュアな接続を試行
			c, err = imapclient.DialInsecure(server, nil)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer c.Logout()

	// 認証情報の取得
	username := ""
	password := ""
	if u.User != nil {
		username = u.User.Username()
		if p, ok := u.User.Password(); ok {
			password = p
		}
	}
	// ログイン
	if err := c.Login(username, password).Wait(); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}
	if pe.Mode == "login" {
		return nil, nil
	}
	// メールボックスの選択 (デフォルトはINBOX)
	mbox := "INBOX"
	if len(u.Path) > 1 {
		mbox = u.Path[1:]
	}
	status, err := c.Select(mbox, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("select failed: %w", err)
	}

	var totalSize int64
	var lastDate time.Time
	var lastID string
	// メッセージが存在する場合、詳細情報を取得
	if status.NumMessages > 0 {
		var seqSet imap.SeqSet
		seqSet.AddRange(1, status.NumMessages)
		fetchOptions := &imap.FetchOptions{RFC822Size: true, InternalDate: true, UID: true}
		messages, err := c.Fetch(seqSet, fetchOptions).Collect()
		if err == nil {
			// 全メッセージのサイズを合算
			for _, msg := range messages {
				totalSize += msg.RFC822Size
			}
			// 最終メッセージの情報を取得
			if len(messages) > 0 {
				lastMsg := messages[len(messages)-1]
				lastDate = lastMsg.InternalDate
				lastID = strconv.FormatUint(uint64(lastMsg.UID), 10)
			}
		}
	}
	// 結果を出力
	return &emailStatsEnt{
		Count:     int(status.NumMessages),
		Size:      totalSize,
		LastDate:  lastDate.Format(time.RFC3339),
		LastMsgID: lastID,
	}, nil
}

// POP3プロトコルによるポーリング処理
func getPOP3Stats(pe *datastore.PollingEnt) (*emailStatsEnt, error) {
	// URLのパース
	u, err := url.Parse(pe.Params)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	server := u.Hostname()
	port := u.Port()
	// ポート番号のデフォルト設定
	if port == "" {
		if u.Scheme == "pop3s" {
			port = "995"
		} else {
			port = "110"
		}
	}
	nport, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	// 認証情報の取得
	username := ""
	password := ""
	if u.User != nil {
		username = u.User.Username()
		if p, ok := u.User.Password(); ok {
			password = p
		}
	}

	// POP3クライアントの設定と接続
	p := pop3.New(pop3.Opt{
		Host:       server,
		Port:       nport,
		TLSEnabled: u.Scheme == "pop3s" || nport == 995,
	})
	conn, err := p.NewConn()
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Quit()

	// 認証
	if err := conn.Auth(username, password); err != nil {
		return nil, fmt.Errorf("auth failed: %w", err)
	}
	if pe.Mode == "login" {
		return nil, nil
	}

	// ステータス情報の取得 (メッセージ数と合計サイズ)
	count, size, err := conn.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat failed: %w", err)
	}

	var lastDate time.Time
	var lastID string
	if count > 0 {
		// 最終メッセージのUIDを取得
		if uidls, err := conn.Uidl(count); err == nil && len(uidls) > 0 {
			lastID = uidls[0].UID
		}
		// 最終メッセージのヘッダーから日付を取得
		if msg, err := conn.Top(count, 0); err == nil {
			dateStr := msg.Header.Get("Date")
			if t, err := mail.ParseDate(dateStr); err == nil {
				lastDate = t
			}
		}
	}

	// 結果を出力
	return &emailStatsEnt{
		Count:     count,
		Size:      int64(size),
		LastDate:  lastDate.Format(time.RFC3339),
		LastMsgID: lastID,
	}, nil
}
