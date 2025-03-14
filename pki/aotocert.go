package pki

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-acme/lego/challenge/http01"
	"github.com/go-acme/lego/challenge/tlsalpn01"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

type AutoCertUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *AutoCertUser) GetEmail() string {
	return u.Email
}
func (u AutoCertUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *AutoCertUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func AutoCert(url, email, sans, dsPath string, insecure bool) error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	autoCertUser := AutoCertUser{
		Email: email,
		key:   privateKey,
	}
	config := lego.NewConfig(&autoCertUser)

	config.CADirURL = url
	config.HTTPClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	client, err := lego.NewClient(config)
	if err != nil {
		return err
	}
	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "80"))
	if err != nil {
		return err
	}
	err = client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer("", "443"))
	if err != nil {
		return err
	}
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return err
	}
	autoCertUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains:    strings.Split(sans, ","),
		Bundle:     true,
		PrivateKey: privateKey,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return err
	}
	os.WriteFile(filepath.Join(dsPath, "key.pem"), certificates.PrivateKey, 0600)
	os.WriteFile(filepath.Join(dsPath, "cert.pem"), certificates.Certificate, 0600)
	return nil
}
