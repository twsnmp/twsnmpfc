package notify

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"

	"github.com/twsnmp/twsnmpfc/datastore"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

var lastState = ""

func GetNotifyOAuth2TokenStep1() (string, error) {
	config := getNotifyOAuth2Config()
	if config == nil {
		return "", fmt.Errorf("no oauth2 config")
	}
	lastState = randCryptoString(32)
	url := config.AuthCodeURL(lastState, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return url, nil
}

func GetNotifyOAuth2TokenStep2(state, code string) error {
	if lastState != "" && state != lastState {
		return fmt.Errorf("state mismatch")
	}
	if code == "" {
		return fmt.Errorf("empty code")
	}
	config := getNotifyOAuth2Config()
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return fmt.Errorf("fail to get token: %w", err)
	}
	datastore.SaveNotifyOAuth2Token(token)
	return nil
}

func getNotifyOAuth2Config() *oauth2.Config {
	redirectURL := fmt.Sprintf("%s/notify/oauth2/callback", datastore.NotifyConf.URL)
	switch datastore.NotifyConf.Provider {
	case "google":
		return &oauth2.Config{
			ClientID:     datastore.NotifyConf.ClientID,
			ClientSecret: datastore.NotifyConf.ClientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  redirectURL,
			Scopes:       []string{"https://mail.google.com/"},
		}
	case "microsoft":
		return &oauth2.Config{
			ClientID:     datastore.NotifyConf.ClientID,
			ClientSecret: datastore.NotifyConf.ClientSecret,
			Endpoint:     microsoft.AzureADEndpoint(datastore.NotifyConf.MSTenant),
			RedirectURL:  redirectURL,
			Scopes:       []string{"https://outlook.office.com/SMTP.Send", "offline_access"},
		}
	default:
		return nil
	}
}

func getNotifyOAuth2Token() *oauth2.Token {
	oldToken := datastore.GetNotifyOAuth2Token()
	if oldToken == nil {
		return nil
	}
	if oldToken.Valid() {
		return oldToken
	}
	config := getNotifyOAuth2Config()
	if config == nil {
		return nil
	}
	tokenSource := config.TokenSource(context.Background(), oldToken)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Printf("Fail to refresh token err=%v", err)
		return nil
	}
	log.Printf("oauth2 token updated old=%v new=%v", oldToken.Expiry, newToken.Expiry)
	datastore.SaveNotifyOAuth2Token(newToken)
	return newToken
}

func randCryptoString(length int) string {
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "randamu_twsnmp"
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
