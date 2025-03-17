package pki

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/twsnmp/twsnmpfc/datastore"
)

var acmeServerPrivateKey []byte
var acmeServerCertificate []byte

var acmeServer *echo.Echo

var lastAcmeServerErr error
var acmeServerRunnning = false

func GetAcmeServerStatus() string {
	if lastAcmeServerErr != nil {
		return fmt.Sprintf("error %v", lastAcmeServerErr)
	} else if acmeServerRunnning {
		return fmt.Sprintf("running port=%d", datastore.PKIConf.AcmePort)
	}
	return "stopped"
}

func startAcmeServer() {
	if acmeServer != nil {
		return
	}
	lastAcmeServerErr = nil
	acmeServerRunnning = true
	acmeServer = echo.New()
	go acmeServerFunc(acmeServer)
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:  time.Now().UnixNano(),
		Type:  "pki",
		Level: "info",
		Event: fmt.Sprintf("ACMEサーバーを起動しました port=%d", datastore.PKIConf.AcmePort),
	})
}

func stopAcmeServer() {
	if acmeServer == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		acmeServer = nil
		lastAcmeServerErr = nil
		acmeServerRunnning = false
	}()
	if err := acmeServer.Shutdown(ctx); err != nil {
		log.Printf("shutdown acme server err=%v", err)
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:  time.Now().UnixNano(),
		Type:  "pki",
		Level: "info",
		Event: "ACMEサーバーを停止しました",
	})

}

func createAcmeServerCertificate() error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	publicKey := &key.PublicKey
	ca, err := x509.ParseCertificate(rootCACertificate)
	if err != nil {
		return err
	}
	sn := getSerial()
	tmp := &x509.Certificate{
		SerialNumber: big.NewInt(sn),
		Subject: pkix.Name{
			CommonName: datastore.PKIConf.Name + " ACME Server",
		},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().AddDate(datastore.PKIConf.RootCATerm, 0, 0).UTC(),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IsCA:                  false,
		BasicConstraintsValid: true,
		CRLDistributionPoints: []string{},
		OCSPServer:            []string{},
	}
	for _, san := range strings.Split(datastore.PKIConf.SANs, ",") {
		baseURL := fmt.Sprintf("http://%s:%d", san, datastore.PKIConf.HttpPort)
		if ip := net.ParseIP(san); ip == nil {
			tmp.DNSNames = append(tmp.DNSNames, san)
		} else {
			tmp.IPAddresses = append(tmp.IPAddresses, ip)
		}
		tmp.CRLDistributionPoints = append(tmp.CRLDistributionPoints, baseURL+"/crl")
		tmp.OCSPServer = append(tmp.OCSPServer, baseURL+"/ocsp")
	}
	if strings.HasPrefix(datastore.PKIConf.HttpBaseURL, "http://") {
		baseURL := strings.TrimRight(datastore.PKIConf.HttpBaseURL, "/")
		tmp.CRLDistributionPoints = append(tmp.CRLDistributionPoints, baseURL+"/crl")
		tmp.OCSPServer = append(tmp.OCSPServer, baseURL+"/ocsp")
	}
	cert, err := x509.CreateCertificate(rand.Reader, tmp, ca, publicKey, rootCAPrivateKey)
	if err != nil {
		return err
	}
	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	acmeServerCertificate = makePEM(cert, "CERTIFICATE")
	acmeServerPrivateKey = makePEM(b, "EC PRIVATE KEY")
	datastore.PKIConf.AcmeServerCert = string(acmeServerCertificate)
	datastore.PKIConf.AcmeServerKey = string(acmeServerPrivateKey)
	datastore.UpdateCert(&datastore.PKICertEnt{
		ID:          fmt.Sprintf("%x", sn),
		Subject:     tmp.Subject.String(),
		Created:     time.Now().UnixNano(),
		Expire:      tmp.NotAfter.UnixNano(),
		Certificate: datastore.PKIConf.AcmeServerCert,
		Type:        "system",
	})
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:  time.Now().UnixNano(),
		Type:  "ca",
		Level: "info",
		Event: fmt.Sprintf("ACMEサーバーの証明証を発行しました subject=%s serial=%x", tmp.Subject.String(), sn),
	})
	return nil
}

// ACME Server
type meta struct {
	TermsOfService          string   `json:"termsOfService,omitempty"`
	Website                 string   `json:"website,omitempty"`
	CaaIdentities           []string `json:"caaIdentities,omitempty"`
	ExternalAccountRequired bool     `json:"externalAccountRequired,omitempty"`
}

// directory represents an ACME directory for configuring clients.
type directory struct {
	NewNonce   string `json:"newNonce"`
	NewAccount string `json:"newAccount"`
	NewOrder   string `json:"newOrder"`
	RevokeCert string `json:"revokeCert"`
	KeyChange  string `json:"keyChange"`
	Meta       *meta  `json:"meta,omitempty"`
}

// externalAccountBinding represents the ACME externalAccountBinding JWS
type externalAccountBinding struct {
	Protected string `json:"protected"`
	Payload   string `json:"payload"`
	Sig       string `json:"signature"`
}

// newAccountRequest represents the payload for a new account request.
type newAccountRequest struct {
	Contact                []string                `json:"contact"`
	OnlyReturnExisting     bool                    `json:"onlyReturnExisting"`
	TermsOfServiceAgreed   bool                    `json:"termsOfServiceAgreed"`
	ExternalAccountBinding *externalAccountBinding `json:"externalAccountBinding,omitempty"`
}

type account struct {
	ID                     string           `json:"-"`
	Key                    *jose.JSONWebKey `json:"-"`
	Contact                []string         `json:"contact,omitempty"`
	Status                 string           `json:"status"`
	OrdersURL              string           `json:"orders"`
	ExternalAccountBinding interface{}      `json:"externalAccountBinding,omitempty"`
	LocationPrefix         string           `json:"-"`
	ProvisionerID          string           `json:"-"`
	ProvisionerName        string           `json:"-"`
}

type acmeCertificate struct {
	ID          string
	AccountID   string
	OrderID     string
	Certificate []byte
}

var accMap = make(map[string]*account)
var orderMap = make(map[string]*order)
var authzMap = make(map[string]*authorization)
var certMap = make(map[string]*acmeCertificate)

type identifier struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type newOrderRequest struct {
	Identifiers []identifier `json:"identifiers"`
	NotBefore   time.Time    `json:"notBefore,omitempty"`
	NotAfter    time.Time    `json:"notAfter,omitempty"`
}

type subproblem struct {
	Type   string `json:"type"`
	Detail string `json:"detail"`
	// The "identifier" field MUST NOT be present at the top level in ACME
	// problem documents.  It can only be present in subproblems.
	// Subproblems need not all have the same type, and they do not need to
	// match the top level type.
	Identifier *identifier `json:"identifier,omitempty"`
}

// acmeError represents an ACME Error
type acmeError struct {
	Type        string       `json:"type"`
	Detail      string       `json:"detail"`
	Subproblems []subproblem `json:"subproblems,omitempty"`
	Err         error        `json:"-"`
	Status      int          `json:"-"`
}

// order contains order metadata for the ACME protocol order type.
type order struct {
	ID                string       `json:"id"`
	AccountID         string       `json:"-"`
	ProvisionerID     string       `json:"-"`
	Status            string       `json:"status"`
	ExpiresAt         time.Time    `json:"expires"`
	Identifiers       []identifier `json:"identifiers"`
	NotBefore         time.Time    `json:"notBefore"`
	NotAfter          time.Time    `json:"notAfter"`
	Error             *acmeError   `json:"error,omitempty"`
	AuthorizationIDs  []string     `json:"-"`
	AuthorizationURLs []string     `json:"authorizations"`
	FinalizeURL       string       `json:"finalize"`
	CertificateID     string       `json:"-"`
	CertificateURL    string       `json:"certificate,omitempty"`
}

type authorization struct {
	ID          string       `json:"-"`
	AccountID   string       `json:"-"`
	Token       string       `json:"-"`
	Fingerprint string       `json:"-"`
	Identifier  identifier   `json:"identifier"`
	Status      string       `json:"status"`
	Challenges  []*challenge `json:"challenges"`
	Wildcard    bool         `json:"wildcard"`
	ExpiresAt   time.Time    `json:"expires"`
	Error       *acmeError   `json:"error,omitempty"`
}

type challenge struct {
	ID              string     `json:"-"`
	AccountID       string     `json:"-"`
	AuthorizationID string     `json:"-"`
	Value           string     `json:"-"`
	Type            string     `json:"type"`
	Status          string     `json:"status"`
	Token           string     `json:"token"`
	ValidatedAt     string     `json:"validated,omitempty"`
	URL             string     `json:"url"`
	Target          string     `json:"target,omitempty"`
	Error           *acmeError `json:"error,omitempty"`
	Payload         []byte     `json:"-"`
	PayloadFormat   string     `json:"-"`
}

type finalizeRequest struct {
	CSR string `json:"csr"`
	csr *x509.CertificateRequest
}

type revokeReqest struct {
	Certificate string `json:"certificate"`
	ReasonCode  *int   `json:"reason,omitempty"`
}

func acmeServerFunc(e *echo.Echo) {
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.GET("/directory", func(c echo.Context) error {
		baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
		d := directory{
			NewNonce:   baseURL + "/new-nonce",
			NewAccount: baseURL + "/new-account",
			NewOrder:   baseURL + "/new-order",
			RevokeCert: baseURL + "/revoke-cert",
			KeyChange:  baseURL + "/key-change",
		}
		return c.JSON(http.StatusOK, d)
	})
	e.HEAD("/new-nonce", func(c echo.Context) error {
		addHeader(c)
		return c.NoContent(http.StatusOK)
	})
	e.GET("/new-nonce", func(c echo.Context) error {
		addHeader(c)
		return c.NoContent(http.StatusNoContent)
	})
	e.POST("/new-account", func(c echo.Context) error {
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("new-account err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		jwk := extractJWK(jws)
		if jwk == nil {
			log.Println("new-account jwk not found")
			return c.JSON(http.StatusBadRequest, err)
		}
		payload, err := jws.Verify(jwk)
		if err != nil {
			log.Printf("new-account err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		var nar newAccountRequest
		if err := json.Unmarshal(payload, &nar); err != nil {
			log.Printf("new-account err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
		// narのチェック、
		acc := &account{
			ID:        jwk.KeyID,
			Key:       jwk,
			Status:    "valid",
			Contact:   nar.Contact,
			OrdersURL: baseURL + "/account/yErbu3KD7bEdxbXsBlaJ82l5VRdtos5W/orders",
		}
		accMap[acc.ID] = acc
		c.Response().Header().Add("Location", baseURL+"/account/"+acc.ID)
		return c.JSON(http.StatusCreated, acc)
	})
	e.POST("/new-order", func(c echo.Context) error {
		baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("new-order err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		acc, err := lookupJWKAndAccount(jws)
		if err != nil {
			log.Printf("new-order err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		payload, err := jws.Verify(acc.Key)
		if err != nil {
			log.Printf("new-order err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		var nor newOrderRequest
		if err := json.Unmarshal(payload, &nor); err != nil {
			log.Printf("new-order err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		now := time.Now().UTC().Truncate(time.Second)
		o := &order{
			ID:                createNonce(),
			AccountID:         acc.ID,
			Status:            "pending",
			Identifiers:       nor.Identifiers,
			ExpiresAt:         now.Add(time.Hour * 24),
			AuthorizationIDs:  make([]string, len(nor.Identifiers)),
			AuthorizationURLs: make([]string, len(nor.Identifiers)),
			NotBefore:         nor.NotBefore,
			NotAfter:          nor.NotAfter,
		}
		for i, identifier := range o.Identifiers {
			azID := createNonce()
			// identifierのタイプ(ホスト名、IPなど)から
			// CAが発行できる証明書であることをチェックし
			// チャレンジのタイプを設定する
			chTypes := []string{}
			switch identifier.Type {
			case "ip":
				chTypes = append(chTypes, "http-01")
				chTypes = append(chTypes, "tls-alpn-01")
			case "dns":
				chTypes = append(chTypes, "dns-01")
				if !strings.HasPrefix(identifier.Value, "*.") {
					chTypes = append(chTypes, "http-01")
					chTypes = append(chTypes, "tls-alpn-01")
				}
			case "permanent-identifier":
				// TPM,Apple,YubiKeyなどのデバイス
				// chTypes = append(chTypes, "device-attest-01")
			case "wireapp-user":
				// OpenID Connect
			case "wireapp-device":
				// OpenID Connect
			}
			az := &authorization{
				ID:         azID,
				AccountID:  acc.ID,
				Identifier: identifier,
				ExpiresAt:  o.ExpiresAt,
				Status:     "pending",
				Challenges: []*challenge{},
				Token:      createNonce(),
			}
			for _, chType := range chTypes {
				chID := createNonce()
				az.Challenges = append(az.Challenges,
					&challenge{
						ID:              chID,
						AccountID:       acc.ID,
						AuthorizationID: azID,
						Type:            chType,
						Value:           identifier.Value,
						Token:           az.Token,
						Status:          "pending",
						URL:             baseURL + "/challenge/" + azID + "/" + chID,
					})
			}
			o.AuthorizationIDs[i] = az.ID
			o.AuthorizationURLs[i] = baseURL + "/authz/" + azID
			authzMap[az.ID] = az
		}
		if o.NotBefore.IsZero() {
			o.NotBefore = now
		}
		if o.NotAfter.IsZero() {
			o.NotAfter = o.NotBefore.Add(time.Hour * 24)
		}
		if nor.NotBefore.IsZero() {
			o.NotBefore = o.NotBefore.Add(-time.Hour * 24)
		}
		o.FinalizeURL = baseURL + "/order/" + o.ID + "/finalize"
		if o.CertificateID != "" {
			o.CertificateURL = baseURL + "/certificate" + o.CertificateID
		}
		orderMap[o.ID] = o
		c.Response().Header().Add("Location", baseURL+"/order/"+o.ID)
		return c.JSON(http.StatusCreated, o)
	})
	e.POST("/authz/:id", func(c echo.Context) error {
		baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("authz err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		id := c.Param("id")
		acc, err := lookupJWKAndAccount(jws)
		if err != nil {
			log.Printf("authz err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		// Check ID
		az, ok := authzMap[id]
		if !ok {
			log.Printf("authz not found id=%s", id)
			return c.JSON(http.StatusBadRequest, err)
		}
		if az.AccountID != acc.ID {
			log.Printf("authz not owner %s!=%s", az.AccountID, acc.ID)
			return c.JSON(http.StatusBadRequest, fmt.Errorf("account id %s!=%s", az.AccountID, acc.ID))
		}
		log.Printf("az=%+v", az)
		for _, c := range az.Challenges {
			log.Printf("challenge=%+v", c)
		}
		c.Response().Header().Add("Location", baseURL+"/authz/"+az.ID)
		return c.JSON(http.StatusOK, az)
	})
	e.POST("/challenge/:authzID/:chID", func(c echo.Context) error {
		baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("challenge err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		authzID := c.Param("authzID")
		chID := c.Param("chID")
		acc, err := lookupJWKAndAccount(jws)
		if err != nil {
			log.Printf("challenge err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		// Check ID
		az, ok := authzMap[authzID]
		if !ok {
			log.Printf("challenge not found id=%s", authzID)
			return c.JSON(http.StatusBadRequest, fmt.Errorf("id not found %s", authzID))
		}
		if az.AccountID != acc.ID {
			log.Printf("challenge acc id mismatch az=%s acc=%s", az.AccountID, acc.ID)
			return c.JSON(http.StatusBadRequest, fmt.Errorf("acount id  mismatch %s!=%s", az.AccountID, acc.ID))
		}
		// Challengeを探す
		var ch *challenge
		for _, c := range az.Challenges {
			if c.ID == chID {
				ch = c
			}
		}
		if ch == nil {
			return c.JSON(http.StatusBadRequest, fmt.Errorf("challenge not found id=%s", chID))
		}
		switch ch.Type {
		case "http-01":
			// ここでHTTPのチェックをしてOKならStatusをvalidにすればよい
			if err := http01Validate(acc, ch); err != nil {
				log.Printf("challenge http01Validate err=%v", err)
				return c.JSON(http.StatusBadRequest, err)
			}
		case "tls-alpn-01":
			if err := tlsalpn01Validate(acc, ch); err != nil {
				log.Printf("challenge tlsalpn01Validate err=%v", err)
				return c.JSON(http.StatusBadRequest, err)
			}
		case "dns-01":
			if err := dns01Validate(acc, ch); err != nil {
				log.Printf("challenge dns01Validate err=%v", err)
				return c.JSON(http.StatusBadRequest, err)
			}
		default:
			return c.JSON(http.StatusBadRequest, fmt.Errorf("challenge not supported type=%s", ch.Type))
		}
		ch.Status = "valid"
		c.Response().Header().Add("Location", baseURL+"/challenge/"+az.ID+"/"+chID)
		c.Response().Header().Add("Link", fmt.Sprintf("<%s/authz/%s>;rel=up", baseURL, az.ID))
		return c.JSON(http.StatusOK, ch)
	})
	e.POST("/order/:id/finalize", func(c echo.Context) error {
		baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		id := c.Param("id")
		acc, err := lookupJWKAndAccount(jws)
		if err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		// Check ID
		o, ok := orderMap[id]
		if !ok {
			log.Printf("finalize not found id=%s", id)
			return c.JSON(http.StatusBadRequest, fmt.Errorf("order not found id=%s", id))
		}
		if o.AccountID != acc.ID {
			log.Printf("finalize acc id mismatch o=%s acc=%s", o.AccountID, acc.ID)
			return c.JSON(http.StatusBadRequest, fmt.Errorf("account id missmatch %s!=%s", o.AccountID, acc.ID))
		}
		payload, err := jws.Verify(acc.Key)
		if err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		var fr finalizeRequest
		if err := json.Unmarshal(payload, &fr); err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		csrBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(fr.CSR, "="))
		if err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		fr.csr, err = x509.ParseCertificateRequest(csrBytes)
		if err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		if err := fr.csr.CheckSignature(); err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		cert, certID, err := createCertificateFromCSR(csrBytes, "acme", map[string]string{"AccountID": o.AccountID, "OrderID": o.ID, "RemoteAddr": c.RealIP()})
		if err != nil {
			log.Printf("finalize err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		pemCert := string(pem.EncodeToMemory(
			&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: cert,
			},
		)) + "\n" +
			string(pem.EncodeToMemory(
				&pem.Block{
					Type:  "CERTIFICATE",
					Bytes: rootCACertificate,
				},
			))
		certMap[certID] = &acmeCertificate{
			ID:          certID,
			OrderID:     o.ID,
			AccountID:   o.AccountID,
			Certificate: []byte(pemCert),
		}
		o.Status = "valid"
		o.CertificateID = certID
		o.CertificateURL = baseURL + "/certificate/" + o.CertificateID
		c.Response().Header().Add("Location", baseURL+"/order/"+id)
		return c.JSON(http.StatusOK, o)
	})
	e.POST("/certificate/:id", func(c echo.Context) error {
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("certificate err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		id := c.Param("id")
		acc, err := lookupJWKAndAccount(jws)
		if err != nil {
			log.Printf("err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		// Check ID
		cert, ok := certMap[id]
		if !ok {
			log.Printf("certificate not found id=%s", id)
			return c.JSON(http.StatusBadRequest, fmt.Errorf("cert not fond id=%s", id))
		}
		if cert.AccountID != acc.ID {
			log.Printf("certificate acc id mismatch o=%s acc=%s", cert.AccountID, acc.ID)
			return c.JSON(http.StatusBadRequest, err)
		}
		return c.Blob(http.StatusOK, "application/pem-certificate-chain", cert.Certificate)
	})
	e.POST("/revoke-cert", func(c echo.Context) error {
		addHeader(c)
		jws, err := getJWS(c)
		if err != nil {
			log.Printf("revoke-cert err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		var jwk *jose.JSONWebKey
		var acc *account
		if canExtractJWKFrom(jws) {
			jwk = extractJWK(jws)
			if jwk == nil {
				log.Println("revoke-cert no jwk")
				return c.JSON(http.StatusBadRequest, fmt.Errorf("missing jwk"))
			}
		} else {
			acc, err = lookupJWKAndAccount(jws)
			if err != nil {
				log.Printf("revoke-cert err=%v", err)
				return c.JSON(http.StatusBadRequest, err)
			}
			jwk = acc.Key
		}
		payload, err := jws.Verify(jwk)
		if err != nil {
			log.Printf("revoke-cert err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		var rvr revokeReqest
		if err := json.Unmarshal(payload, &rvr); err != nil {
			log.Printf("revoke-cert err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		// nvrのチェック、
		certBytes, err := base64.RawURLEncoding.DecodeString(rvr.Certificate)
		if err != nil {
			log.Printf("revoke-cert err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		certToBeRevoked, err := x509.ParseCertificate(certBytes)
		if err != nil {
			log.Printf("revoke-cert err=%v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		serial := certToBeRevoked.SerialNumber.Int64()

		cert := datastore.FindCert(fmt.Sprintf("%x", serial))
		if cert == nil {
			log.Printf("revoke-cert not found serial=%x", serial)
			return c.JSON(http.StatusBadRequest, fmt.Errorf("cert not found serial=%x", serial))
		}
		// ここで失効できるかチェックする
		// jwkが送信された場合とアカウントから取得した場合でチェック方法が変わる
		if acc != nil {
			// アカウントの場合は、登録しているアカウントが証明書の所有者
			accID, ok := cert.Info["AccountID"]
			if !ok || acc.ID != accID {
				log.Printf("revoke-cert not owner serial=%x", serial)
				return c.JSON(http.StatusBadRequest, fmt.Errorf("you do not owwn this cert serial=%x", serial))
			}
		} else {
			// 失効する証明書の秘密鍵で検証する
			if _, err := jws.Verify(certToBeRevoked.PublicKey); err != nil {
				log.Printf("revoke-cert err=%v", err)
				return c.JSON(http.StatusBadRequest, err)
			}
		}
		datastore.RevokeCert(cert)
		return c.NoContent(http.StatusOK)
	})
	if err := e.StartTLS(fmt.Sprintf(":%d", datastore.PKIConf.AcmePort), acmeServerCertificate, acmeServerPrivateKey); err != http.ErrServerClosed {
		acmeServerRunnning = false
		lastAcmeServerErr = err
		log.Printf("acme server startTLS err=%v", err)
	}
}

func addHeader(c echo.Context) {
	baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
	c.Response().Header().Add("Replay-Nonce", createNonce())
	c.Response().Header().Add("Link", baseURL+"/directory/index")
	c.Response().Header().Add(echo.HeaderCacheControl, "no-store")
}

func getJWS(c echo.Context) (*jose.JSONWebSignature, error) {
	baseURL := strings.TrimRight(datastore.PKIConf.AcmeBaseURL, "/")
	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return nil, err
	}
	jws, err := jose.ParseSigned(string(b), []jose.SignatureAlgorithm{jose.ES256, jose.ES384})
	if err != nil {
		return nil, err
	}
	if err := checkJWS(jws, baseURL); err != nil {
		return nil, err
	}
	return jws, err
}

func createNonce() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "bad nonce"
	}
	var s string
	for _, v := range b {
		s += string(v%byte(94) + 33)
	}
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

func http01Validate(acc *account, ch *challenge) error {
	u := &url.URL{Scheme: "http", Host: ch.Value, Path: fmt.Sprintf("/.well-known/acme-challenge/%s", ch.Token)}
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	r, err := client.Get(u.String())
	if err != nil {
		return err
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	keyAuth := strings.TrimSpace(string(body))
	expected, err := keyAuthorization(ch.Token, acc.Key)
	if err != nil {
		return err
	}
	if keyAuth != expected {
		return fmt.Errorf("http get %s != %s", keyAuth, expected)
	}
	return nil
}

func serverName(ch *challenge) string {
	if ip := net.ParseIP(ch.Value); ip != nil {
		return reverseAddr(ip)
	}
	return ch.Value
}

func reverseAddr(ip net.IP) (arpa string) {
	if ip.To4() != nil {
		return uitoa(uint(ip[15])) + "." + uitoa(uint(ip[14])) + "." + uitoa(uint(ip[13])) + "." + uitoa(uint(ip[12])) + ".in-addr.arpa."
	}
	// Must be IPv6
	buf := make([]byte, 0, len(ip)*4+len("ip6.arpa."))
	// Add it, in reverse, to the buffer
	for i := len(ip) - 1; i >= 0; i-- {
		v := ip[i]
		buf = append(buf, hexit[v&0xF],
			'.',
			hexit[v>>4],
			'.')
	}
	// Append "ip6.arpa." and return (buf already has the final .)
	buf = append(buf, "ip6.arpa."...)
	return string(buf)
}

const hexit = "0123456789abcdef"

func uitoa(val uint) string {
	if val == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf) - 1
	for val >= 10 {
		v := val / 10
		buf[i] = byte('0' + val - v*10)
		i--
		val = v
	}
	buf[i] = byte('0' + val)
	return string(buf[i:])
}

func tlsalpn01Validate(acc *account, ch *challenge) error {
	config := &tls.Config{
		NextProtos:         []string{"acme-tls/1"},
		MinVersion:         tls.VersionTLS12,
		ServerName:         serverName(ch),
		InsecureSkipVerify: true,
	}
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}
	hostPort := net.JoinHostPort(ch.Value, "443")

	conn, err := tls.DialWithDialer(dialer, "tcp", hostPort, config)
	if err != nil {
		return err
	}
	defer conn.Close()

	cs := conn.ConnectionState()
	certs := cs.PeerCertificates

	if len(certs) == 0 {
		return fmt.Errorf("certificate not found for %s", hostPort)
	}
	if cs.NegotiatedProtocol != "acme-tls/1" {
		return fmt.Errorf("cannot negotiate ALPN acme-tls/1 protocol for tls-alpn-01 challenge")
	}

	leafCert := certs[0]

	// if no DNS names present, look for IP address and verify that exactly one exists
	if len(leafCert.DNSNames) == 0 {
		if len(leafCert.IPAddresses) != 1 || !leafCert.IPAddresses[0].Equal(net.ParseIP(ch.Value)) {
			return fmt.Errorf("incorrect certificate for tls-alpn-01 challenge: leaf certificate must contain a single IP address or DNS name, %v", ch.Value)
		}
	} else {
		if len(leafCert.DNSNames) != 1 || !strings.EqualFold(leafCert.DNSNames[0], ch.Value) {
			return fmt.Errorf("incorrect certificate for tls-alpn-01 challenge: leaf certificate must contain a single IP address or DNS name, %v", ch.Value)
		}
	}

	idPeAcmeIdentifier := asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 31}
	idPeAcmeIdentifierV1Obsolete := asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 30, 1}
	foundIDPeAcmeIdentifierV1Obsolete := false

	keyAuth, err := keyAuthorization(ch.Token, acc.Key)
	if err != nil {
		return err
	}
	hashedKeyAuth := sha256.Sum256([]byte(keyAuth))

	for _, ext := range leafCert.Extensions {
		if idPeAcmeIdentifier.Equal(ext.Id) {
			if !ext.Critical {
				return fmt.Errorf("incorrect certificate for tls-alpn-01 challenge: acmeValidationV1 extension not critical")
			}
			var extValue []byte
			rest, err := asn1.Unmarshal(ext.Value, &extValue)

			if err != nil || len(rest) > 0 || len(hashedKeyAuth) != len(extValue) {
				return fmt.Errorf("incorrect certificate for tls-alpn-01 challenge: malformed acmeValidationV1 extension value")
			}
			if subtle.ConstantTimeCompare(hashedKeyAuth[:], extValue) != 1 {
				return fmt.Errorf("incorrect certificate for tls-alpn-01 challenge: expected acmeValidationV1 extension value %s for this challenge but got %s",
					hex.EncodeToString(hashedKeyAuth[:]), hex.EncodeToString(extValue))
			}
			ch.Status = "valid"
			ch.Error = nil
			ch.ValidatedAt = time.Now().Truncate(time.Second).Format(time.RFC3339)
			return nil
		}
		if idPeAcmeIdentifierV1Obsolete.Equal(ext.Id) {
			foundIDPeAcmeIdentifierV1Obsolete = true
		}
	}

	if foundIDPeAcmeIdentifierV1Obsolete {
		return fmt.Errorf("incorrect certificate for tls-alpn-01 challenge: obsolete id-pe-acmeIdentifier in acmeValidationV1 extension")
	}
	return fmt.Errorf("incorrect certificate for tls-alpn-01 challenge: missing acmeValidationV1 extension")
}

func dns01Validate(acc *account, ch *challenge) error {
	// Normalize domain for wildcard DNS names
	// This is done to avoid making TXT lookups for domains like
	// _acme-challenge.*.example.com
	// Instead perform txt lookup for _acme-challenge.example.com
	domain := strings.TrimPrefix(ch.Value, "*.")

	txtRecords, err := net.LookupTXT("_acme-challenge." + domain)
	if err != nil {
		return fmt.Errorf("error looking up TXT records for domain %s", domain)
	}

	expectedKeyAuth, err := keyAuthorization(ch.Token, acc.Key)
	if err != nil {
		return err
	}
	h := sha256.Sum256([]byte(expectedKeyAuth))
	expected := base64.RawURLEncoding.EncodeToString(h[:])
	var found bool
	for _, r := range txtRecords {
		if r == expected {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("keyAuthorization does not match; expected %s, but got %s", expectedKeyAuth, txtRecords)
	}
	// Update and store the challenge.
	ch.Status = "valid"
	ch.Error = nil
	ch.ValidatedAt = time.Now().Truncate(time.Second).Format(time.RFC3339)
	return nil
}

// keyAuthorization creates the ACME key authorization value from a token
// and a jwk.
func keyAuthorization(token string, jwk *jose.JSONWebKey) (string, error) {
	thumbprint, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", err
	}
	encPrint := base64.RawURLEncoding.EncodeToString(thumbprint)
	return fmt.Sprintf("%s.%s", token, encPrint), nil
}

func canExtractJWKFrom(jws *jose.JSONWebSignature) bool {
	if jws == nil {
		return false
	}
	if len(jws.Signatures) == 0 {
		return false
	}
	return jws.Signatures[0].Protected.JSONWebKey != nil
}

// JWSのチェック

// validateJWS checks the request body for to verify that it meets ACME
// requirements for a JWS.
//
// The JWS MUST NOT have multiple signatures
// The JWS Unencoded Payload Option [RFC7797] MUST NOT be used
// The JWS Unprotected Header [RFC7515] MUST NOT be used
// The JWS Payload MUST NOT be detached
// The JWS Protected Header MUST include the following fields:
//   - “alg” (Algorithm).
//     This field MUST NOT contain “none” or a Message Authentication Code
//     (MAC) algorithm (e.g. one in which the algorithm registry description
//     mentions MAC/HMAC).
//   - “nonce” (defined in Section 6.5)
//   - “url” (defined in Section 6.4)
//   - Either “jwk” (JSON Web Key) or “kid” (Key ID) as specified below<Paste>
func checkJWS(jws *jose.JSONWebSignature, baseURL string) error {
	if len(jws.Signatures) == 0 {
		return fmt.Errorf("request body does not contain a signature")
	}
	if len(jws.Signatures) > 1 {
		return fmt.Errorf("request body contains more than one signature")
	}
	sig := jws.Signatures[0]
	uh := sig.Unprotected
	if uh.KeyID != "" ||
		uh.JSONWebKey != nil ||
		uh.Algorithm != "" ||
		uh.Nonce != "" ||
		len(uh.ExtraHeaders) > 0 {
		return fmt.Errorf("unprotected header must not be used")
	}
	// Algorithmのチェックは不要
	hdr := sig.Protected
	// Check the validity/freshness of the Nonce.
	if hdr.Nonce == "" {
		return fmt.Errorf("invalid nonce")
	}
	// Check that the JWS url matches the requested url.
	if jwsURL, ok := hdr.ExtraHeaders["url"].(string); !ok || !strings.HasPrefix(jwsURL, baseURL) {
		return fmt.Errorf("jws missing url protected header")
	}
	if hdr.JSONWebKey != nil && hdr.KeyID != "" {
		return fmt.Errorf("jwk and kid are mutually exclusive")
	}
	if hdr.JSONWebKey == nil && hdr.KeyID == "" {
		return fmt.Errorf("either jwk or kid must be defined in jws protected header")
	}
	return nil
}

// JWKを取り出す
func extractJWK(jws *jose.JSONWebSignature) *jose.JSONWebKey {
	jwk := jws.Signatures[0].Protected.JSONWebKey
	if jwk == nil {
		log.Println("jwk expected in protected header")
		return nil
	}
	if !jwk.Valid() {
		log.Println("invalid jwk in protected header")
		return nil
	}
	kid, err := keyToID(jwk)
	if err == nil {
		jwk.KeyID = kid
	}
	return jwk

}

func lookupJWKAndAccount(jws *jose.JSONWebSignature) (*account, error) {
	kid := jws.Signatures[0].Protected.KeyID
	if kid == "" {
		return nil, fmt.Errorf("jws missing keyid")
	}
	kid = path.Base(kid)
	acc, ok := accMap[kid]
	if !ok {
		return nil, fmt.Errorf("no account for %s", kid)
	}
	return acc, nil
}

// keyToID converts a JWK to a thumbprint.
func keyToID(jwk *jose.JSONWebKey) (string, error) {
	kid, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(kid), nil
}
