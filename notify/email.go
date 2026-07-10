// Package notify : 通知処理
package notify

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/wneessen/go-mail"

	"github.com/twsnmp/twsnmpfc/datastore"
)

func canSendMail() bool {
	if datastore.NotifyConf.MailFrom == "" ||
		datastore.NotifyConf.MailTo == "" {
		return false
	}
	switch datastore.NotifyConf.Provider {
	case "google", "microsoft":
		return datastore.HasValidNotifyOAuth2Token(&datastore.NotifyConf)
	default:
		if datastore.NotifyConf.MailServer == "" {
			return false
		}
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
		level := "info"
		if err != nil {
			log.Printf("send mail err=%v", err)
			r = fmt.Sprintf("失敗 エラー=%v", err)
			level = "high"
		}
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: level,
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

// SendActionMail sends a notification email with the specified subject and body.
func SendActionMail(subject, body string) {
	if err := sendMail(subject, body+"\r\n"+datastore.NotifyConf.URL); err != nil {
		log.Printf("SendActionMail err=%v", err)
	}
}

func sendMail(subject, body string) error {
	if !canSendMail() {
		return nil
	}
	switch datastore.NotifyConf.Provider {
	case "google":
		return sendMailOAuth2("smtp.gmail.com", subject, body)
	case "microsoft":
		return sendMailOAuth2("smtp-mail.outlook.com", subject, body)
	default:
		return sendMailSMTP(subject, body)
	}
}

func sendMailSMTP(subject, body string) error {
	host, portStr, err := net.SplitHostPort(datastore.NotifyConf.MailServer)
	var port int
	if err != nil {
		host = datastore.NotifyConf.MailServer
	} else {
		port, _ = strconv.Atoi(portStr)
	}
	tlsconfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: datastore.NotifyConf.InsecureSkipVerify,
	}
	if datastore.NotifyConf.InsecureCipherSuites {
		for _, e := range tls.CipherSuites() {
			tlsconfig.CipherSuites = append(tlsconfig.CipherSuites, e.ID)
		}
		tlsconfig.CipherSuites = append(tlsconfig.CipherSuites, tls.TLS_RSA_WITH_AES_128_GCM_SHA256)
		tlsconfig.CipherSuites = append(tlsconfig.CipherSuites, tls.TLS_RSA_WITH_AES_256_GCM_SHA384)
	}

	opts := []mail.Option{
		mail.WithTLSConfig(tlsconfig),
	}
	if port > 0 {
		opts = append(opts, mail.WithPort(port))
	}
	if strings.HasSuffix(datastore.NotifyConf.MailServer, ":465") {
		opts = append(opts, mail.WithSSL())
	} else {
		opts = append(opts, mail.WithTLSPortPolicy(mail.TLSOpportunistic))
	}
	if datastore.NotifyConf.User != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
			mail.WithUsername(datastore.NotifyConf.User),
			mail.WithPassword(datastore.NotifyConf.Password),
		)
	}

	client, err := mail.NewClient(host, opts...)
	if err != nil {
		log.Printf("send mail err=%v", err)
		return err
	}

	message := mail.NewMsg()
	if err := message.From(datastore.NotifyConf.MailFrom); err != nil {
		log.Printf("send mail err=%v", err)
		return err
	}
	for _, rcpt := range strings.Split(datastore.NotifyConf.MailTo, ",") {
		if !strings.Contains(rcpt, "@") {
			continue
		}
		if err := message.AddTo(rcpt); err != nil {
			log.Printf("send mail err=%v", err)
			return err
		}
	}

	message.Subject(subject)
	if datastore.NotifyConf.HTMLMail {
		message.SetBodyString(mail.TypeTextHTML, body)
	} else {
		message.SetBodyString(mail.TypeTextPlain, body)
	}

	if err := client.DialAndSend(message); err != nil {
		log.Printf("send mail err=%v", err)
		return err
	}
	log.Printf("send mail to %s", datastore.NotifyConf.MailTo)
	return nil
}

func SendTestMail(testConf *datastore.NotifyConfEnt) error {
	switch testConf.Provider {
	case "google":
		return sendTestMailOAuth2("smtp.gmail.com", testConf)
	case "microsoft":
		return sendTestMailOAuth2("smtp-mail.outlook.com", testConf)
	default:
		return sendTestMailSMTP(testConf)
	}
}

func sendTestMailSMTP(testConf *datastore.NotifyConfEnt) error {
	host, portStr, err := net.SplitHostPort(testConf.MailServer)
	var port int
	if err != nil {
		host = testConf.MailServer
	} else {
		port, _ = strconv.Atoi(portStr)
	}
	tlsconfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: testConf.InsecureSkipVerify,
	}
	if testConf.InsecureCipherSuites {
		for _, e := range tls.CipherSuites() {
			tlsconfig.CipherSuites = append(tlsconfig.CipherSuites, e.ID)
		}
		tlsconfig.CipherSuites = append(tlsconfig.CipherSuites, tls.TLS_RSA_WITH_AES_128_GCM_SHA256)
		tlsconfig.CipherSuites = append(tlsconfig.CipherSuites, tls.TLS_RSA_WITH_AES_256_GCM_SHA384)
	}

	opts := []mail.Option{
		mail.WithTLSConfig(tlsconfig),
	}
	if port > 0 {
		opts = append(opts, mail.WithPort(port))
	}
	if strings.HasSuffix(testConf.MailServer, ":465") {
		opts = append(opts, mail.WithSSL())
	} else {
		opts = append(opts, mail.WithTLSPortPolicy(mail.TLSOpportunistic))
	}
	if testConf.User != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
			mail.WithUsername(testConf.User),
			mail.WithPassword(testConf.Password),
		)
	}

	client, err := mail.NewClient(host, opts...)
	if err != nil {
		log.Printf("send test mail err=%v", err)
		return err
	}

	message := mail.NewMsg()
	if err := message.From(testConf.MailFrom); err != nil {
		log.Printf("send test mail err=%v", err)
		return err
	}
	for _, rcpt := range strings.Split(testConf.MailTo, ",") {
		if !strings.Contains(rcpt, "@") {
			continue
		}
		if err := message.AddTo(rcpt); err != nil {
			log.Printf("send test mail err=%v", err)
			return err
		}
	}

	message.Subject(testConf.Subject)
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
		message.SetBodyString(mail.TypeTextHTML, buffer.String())
	} else {
		message.SetBodyString(mail.TypeTextPlain, "Test Mail.\r\n試験メール.\r\n")
	}

	if err := client.DialAndSend(message); err != nil {
		log.Printf("send test mail err=%v", err)
		return err
	}
	return nil
}

func sendMailOAuth2(server, subject, body string) error {
	token := getNotifyOAuth2Token()
	if token == nil {
		return fmt.Errorf("oauth2 token not found")
	}
	client, err := mail.NewClient(server,
		mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithSMTPAuth(mail.SMTPAuthXOAUTH2),
		mail.WithUsername(datastore.NotifyConf.User), mail.WithPassword(token.AccessToken))
	if err != nil {
		return err
	}
	message := mail.NewMsg()
	if err := message.From(datastore.NotifyConf.MailFrom); err != nil {
		return err
	}
	for _, rcpt := range strings.Split(datastore.NotifyConf.MailTo, ",") {
		if !strings.Contains(rcpt, "@") {
			continue
		}
		if err := message.AddTo(rcpt); err != nil {
			return err
		}
	}

	message.Subject(subject)
	if datastore.NotifyConf.HTMLMail {
		message.SetBodyString(mail.TypeTextHTML, body)
	} else {
		message.SetBodyString(mail.TypeTextPlain, body)
	}
	return client.DialAndSend(message)
}

func sendTestMailOAuth2(server string, testConf *datastore.NotifyConfEnt) error {
	token := getNotifyOAuth2Token()
	if token == nil {
		return fmt.Errorf("oauth2 token not found")
	}
	log.Printf("send test mail token=%v", token.Expiry)
	client, err := mail.NewClient(server,
		mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithSMTPAuth(mail.SMTPAuthXOAUTH2),
		mail.WithUsername(testConf.User), mail.WithPassword(token.AccessToken))
	if err != nil {
		return err
	}
	message := mail.NewMsg()
	if err := message.From(testConf.MailFrom); err != nil {
		return err
	}
	for _, rcpt := range strings.Split(testConf.MailTo, ",") {
		if !strings.Contains(rcpt, "@") {
			continue
		}
		if err := message.AddTo(rcpt); err != nil {
			return err
		}
	}
	message.Subject(testConf.Subject)
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
		message.SetBodyString(mail.TypeTextHTML, buffer.String())
	} else {
		message.SetBodyString(mail.TypeTextPlain, "Test Mail.\r\n試験メール.\r\n")
	}
	return client.DialAndSend(message)
}
