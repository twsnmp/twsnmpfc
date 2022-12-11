// Package notify : 通知処理
package notify

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/smtp"
	"strings"

	"github.com/twsnmp/twsnmpfc/datastore"
)

func canSendMail() bool {
	if datastore.NotifyConf.MailServer == "" ||
		datastore.NotifyConf.MailFrom == "" ||
		datastore.NotifyConf.MailTo == "" {
		return false
	}
	return true
}

func sendNotifyMail(list []*datastore.EventLogEnt) {
	if !canSendMail() {
		return
	}
	nl := getLevelNum(datastore.NotifyConf.Level)
	if nl == 3 {
		return
	}
	nd := getNotifyData(list, nl)
	if nd.failureBody != "" {
		err := sendMail(nd.failureSubject, nd.failureBody)
		r := ""
		if err != nil {
			log.Printf("send mail err=%v", err)
			r = fmt.Sprintf("失敗 エラー=%v", err)
		}
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: fmt.Sprintf("通知メール送信 %s", r),
		})
	}
	if nd.repairBody != "" {
		err := sendMail(nd.repairSubject, nd.repairBody)
		r := ""
		if err != nil {
			log.Printf("send mail err=%v", err)
			r = fmt.Sprintf("失敗 エラー=%v", err)
		}
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: fmt.Sprintf("復帰通知メール送信 %s", r),
		})
	}
}

func sendMail(subject, body string) error {
	if !canSendMail() {
		return nil
	}
	host, _, err := net.SplitHostPort(datastore.NotifyConf.MailServer)
	if err != nil {
		host = datastore.NotifyConf.MailServer
	}
	tlsconfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: datastore.NotifyConf.InsecureSkipVerify,
	}
	var c *smtp.Client
	if strings.HasSuffix(datastore.NotifyConf.MailServer, ":465") {
		conn, err := tls.Dial("tcp", datastore.NotifyConf.MailServer, tlsconfig)
		if err != nil {
			log.Printf("send mail err=%v", err)
			return err
		}
		c, err = smtp.NewClient(conn, host)
		if err != nil {
			log.Printf("send mail err=%v", err)
			return err
		}
	} else {
		c, err = smtp.Dial(datastore.NotifyConf.MailServer)
		if err != nil {
			return err
		}
		if err = c.StartTLS(tlsconfig); err != nil {
			log.Printf("send mail err=%s", err)
		}
	}
	defer c.Close()
	if datastore.NotifyConf.User != "" {
		auth := smtp.PlainAuth("", datastore.NotifyConf.User, datastore.NotifyConf.Password, host)
		if err = c.Auth(auth); err != nil {
			log.Printf("send mail err=%s", err)
			return err
		}
	}
	if err = c.Mail(datastore.NotifyConf.MailFrom); err != nil {
		log.Printf("send mail err=%s", err)
		return err
	}
	for _, rcpt := range strings.Split(datastore.NotifyConf.MailTo, ",") {
		if err = c.Rcpt(rcpt); err != nil {
			log.Printf("send mail err=%s", err)
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		log.Printf("send mail err=%s", err)
		return err
	}
	defer w.Close()
	body = convNewline(body, "\r\n")
	message := makeMailMessage(datastore.NotifyConf.MailFrom, datastore.NotifyConf.MailTo, subject, body, datastore.NotifyConf.HTMLMail)
	_, _ = w.Write([]byte(message))
	_ = c.Quit()
	log.Printf("send mail to %s", datastore.NotifyConf.MailTo)
	return nil
}

func convNewline(str, nlcode string) string {
	return strings.NewReplacer(
		"\r\n", nlcode,
		"\r", nlcode,
		"\n", nlcode,
	).Replace(str)
}

func SendTestMail(testConf *datastore.NotifyConfEnt) error {
	host, _, err := net.SplitHostPort(testConf.MailServer)
	if err != nil {
		host = testConf.MailServer
	}
	tlsconfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: testConf.InsecureSkipVerify,
	}
	var c *smtp.Client
	if strings.HasSuffix(testConf.MailServer, ":465") {
		conn, err := tls.Dial("tcp", testConf.MailServer, tlsconfig)
		if err != nil {
			log.Printf("send test mail err=%v", err)
			return err
		}
		c, err = smtp.NewClient(conn, host)
		if err != nil {
			log.Printf("send test mail err=%v", err)
			return err
		}
	} else {
		c, err = smtp.Dial(testConf.MailServer)
		if err != nil {
			log.Printf("send test mail err=%s", err)
			return err
		}
		if err = c.StartTLS(tlsconfig); err != nil {
			log.Printf("send test mail err=%s", err)
		}
	}
	defer c.Close()
	if testConf.User != "" {
		auth := smtp.PlainAuth("", testConf.User, testConf.Password, host)
		if err = c.Auth(auth); err != nil {
			log.Printf("send test mail err=%s", err)
			return err
		}
	}
	if err = c.Mail(testConf.MailFrom); err != nil {
		log.Printf("send test mail err=%s", err)
		return err
	}
	for _, rcpt := range strings.Split(testConf.MailTo, ",") {
		if err = c.Rcpt(rcpt); err != nil {
			log.Printf("send test mail err=%s", err)
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		log.Printf("send test mail err=%s", err)
		return err
	}
	defer w.Close()
	body := "Test Mail.\r\n試験メール.\r\n"
	if testConf.HTMLMail {
		t, err := template.New("test").Parse(datastore.LoadMailTemplate("test"))
		if err != nil {
			log.Printf("send test mail err=%s", err)
			return err
		}
		buffer := new(bytes.Buffer)
		if err = t.Execute(buffer, map[string]interface{}{
			"Title": testConf.Subject + "(試験メール）",
			"URL":   testConf.URL,
		}); err != nil {
			return err
		}
		body = buffer.String()
	}
	message := makeMailMessage(testConf.MailFrom, testConf.MailTo, testConf.Subject, body, testConf.HTMLMail)
	_, _ = w.Write([]byte(message))
	_ = c.Quit()
	return nil
}

func makeMailMessage(from, to, subject, body string, bHTML bool) string {
	var header bytes.Buffer
	header.WriteString("From: " + from + "\r\n")
	header.WriteString("To: " + to + "\r\n")
	header.WriteString(encodeSubject(subject))
	header.WriteString("MIME-Version: 1.0\r\n")
	if bHTML {
		header.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
		var message bytes.Buffer = header
		message.WriteString("\r\n")
		message.WriteString(body)
		return message.String()
	}
	header.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	header.WriteString("Content-Transfer-Encoding: base64\r\n")

	var message bytes.Buffer = header
	message.WriteString("\r\n")
	message.WriteString(add76crlf(base64.StdEncoding.EncodeToString([]byte(body))))

	return message.String()
}

// 76バイト毎にCRLFを挿入する
func add76crlf(msg string) string {
	var buffer bytes.Buffer
	for k, c := range strings.Split(msg, "") {
		buffer.WriteString(c)
		if k%76 == 75 {
			buffer.WriteString("\r\n")
		}
	}
	return buffer.String()
}

// UTF8文字列を指定文字数で分割
func utf8Split(utf8string string, length int) []string {
	resultString := []string{}
	var buffer bytes.Buffer
	for k, c := range strings.Split(utf8string, "") {
		buffer.WriteString(c)
		if k%length == length-1 {
			resultString = append(resultString, buffer.String())
			buffer.Reset()
		}
	}
	if buffer.Len() > 0 {
		resultString = append(resultString, buffer.String())
	}
	return resultString
}

// サブジェクトをMIMEエンコードする
func encodeSubject(subject string) string {
	var buffer bytes.Buffer
	buffer.WriteString("Subject:")
	for _, line := range utf8Split(subject, 13) {
		buffer.WriteString(" =?utf-8?B?")
		buffer.WriteString(base64.StdEncoding.EncodeToString([]byte(line)))
		buffer.WriteString("?=\r\n")
	}
	return buffer.String()
}
