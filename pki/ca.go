// Package pki provides functions for managing a public key infrastructure (PKI),
// including certificate authority (CA) operations, certificate issuance, and revocation.
package pki

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

var rootCAPrivateKey any
var rootCAPublicKey any
var rootCACertificate []byte
var crl []byte

func CreateCA(req *datastore.CreateCAReq) error {
	datastore.InitCAConf(req)
	if err := createRootCACertificate(); err != nil {
		return err
	}
	createScepCACertificate()
	createAcmeServerCertificate()
	return datastore.SavePKIConf()
}

var stopCA = false

func DestroyCA() {
	log.Println("destroy CA")
	datastore.ClearCAData()
	stopCA = true
}

func Start(ctx context.Context, wg *sync.WaitGroup) error {
	if err := loadRootCA(); err == nil {
		if datastore.PKIConf.AcmeServerKey != "" && datastore.PKIConf.AcmeServerCert != "" {
			acmeServerPrivateKey = []byte(datastore.PKIConf.AcmeServerKey)
			acmeServerCertificate = []byte(datastore.PKIConf.AcmeServerCert)
		}
		loadScepCA()
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:  time.Now().UnixNano(),
		Type:  "pki",
		Level: "info",
		Event: "PKI機能を開始しました",
	})
	wg.Add(1)
	go caServer(ctx, wg)
	return nil
}
func caServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start ca server")
	timer := time.NewTicker(time.Second * 1)
	lastCrlTime := int64(6)
	for {
		select {
		case <-ctx.Done():
			stopAcmeServer()
			stopHTTPServer()
			return
		case <-timer.C:
			if rootCAPrivateKey == nil {
				continue
			}
			if datastore.PKIConf.EnableAcme && acmeServer == nil {
				log.Printf("start acme server port=%d", datastore.PKIConf.AcmePort)
				startAcmeServer()
			} else if !datastore.PKIConf.EnableAcme && acmeServer != nil {
				log.Printf("stop acme server port=%d", datastore.PKIConf.AcmePort)
				stopAcmeServer()
			}
			if datastore.PKIConf.EnableHTTP && httpServer == nil {
				log.Printf("start http server port=%d", datastore.PKIConf.HTTPPort)
				startHTTPServer()
			} else if !datastore.PKIConf.EnableHTTP && httpServer != nil {
				log.Printf("stop http server port=%d", datastore.PKIConf.HTTPPort)
				stopHTTPServer()
			}
			now := time.Now().Unix()
			if now-lastCrlTime > int64(datastore.PKIConf.CrlInterval*3600) && IsCAValid() {
				createCRL()
				lastCrlTime = now
			}
			if stopCA {
				log.Println("clear CA data")
				rootCAPrivateKey = nil
				rootCAPublicKey = nil
				rootCACertificate = nil
				crl = nil
				stopCA = false
			}
		}
	}
}

func IsCAValid() bool {
	return rootCACertificate != nil && rootCAPrivateKey != nil && !stopCA
}

type CSRReqEnt struct {
	KeyType            string `json:"KeyType"`
	CommonName         string `json:"CommonName"`
	OrganizationalUnit string `json:"OrganizationalUnit"`
	Organization       string `json:"Organization"`
	Locality           string `json:"Locality"`
	Province           string `json:"Province"`
	Country            string `json:"Country"`
	Sans               string `json:"Sans"`
}

func CreateCertificateRequest(req *CSRReqEnt) ([]byte, error) {
	var key any
	var keyBytes []byte
	var err error
	tmp := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: req.CommonName,
		},
	}
	if req.Country != "" {
		tmp.Subject.Country = append(tmp.Subject.Country, req.Country)
	}
	if req.Province != "" {
		tmp.Subject.Province = append(tmp.Subject.Province, req.Province)
	}
	if req.Locality != "" {
		tmp.Subject.Locality = append(tmp.Subject.Locality, req.Locality)
	}
	if req.Organization != "" {
		tmp.Subject.Organization = append(tmp.Subject.Organization, req.Organization)
	}
	if req.OrganizationalUnit != "" {
		tmp.Subject.OrganizationalUnit = append(tmp.Subject.OrganizationalUnit, req.OrganizationalUnit)
	}
	for _, san := range strings.Split(req.Sans, ",") {
		san = strings.TrimSpace(san)
		if san == "" {
			continue
		}
		if ip := net.ParseIP(san); ip != nil {
			tmp.IPAddresses = append(tmp.IPAddresses, ip)
		} else if strings.Contains(san, "@") {
			tmp.EmailAddresses = append(tmp.EmailAddresses, san)
		} else {
			tmp.DNSNames = append(tmp.DNSNames, san)
		}
	}
	pemKeyType := "RSA PRIVATE KEY"
	if strings.HasPrefix(req.KeyType, "rsa-") {
		bits := 4096
		switch req.KeyType {
		case "rsa-2048":
			bits = 2048
		case "rsa-8192":
			bits = 8192
		}
		k, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			return nil, err
		}
		tmp.PublicKeyAlgorithm = x509.RSA
		tmp.SignatureAlgorithm = x509.SHA256WithRSA
		tmp.PublicKey = &k.PublicKey
		key = k
	} else {
		var curve = elliptic.P256()
		switch req.KeyType {
		case "ecdsa-224":
			curve = elliptic.P224()
		case "ecdsa-384":
			curve = elliptic.P384()
		case "ecdsa-521":
			curve = elliptic.P521()
		}
		k, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			return nil, err
		}
		tmp.PublicKeyAlgorithm = x509.ECDSA
		tmp.SignatureAlgorithm = x509.ECDSAWithSHA256
		tmp.PublicKey = &k.PublicKey
		key = k
		pemKeyType = "EC PRIVATE KEY"
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, tmp, key)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create("csr.pem")
	if err != nil {
		return nil, err
	}
	_, err = f.Write(makePEM(csr, "CERTIFICATE REQUEST"))
	if err != nil {
		return nil, err
	}
	f, err = w.Create("key.pem")
	if err != nil {
		return nil, err
	}
	_, err = f.Write(makePEM(keyBytes, pemKeyType))
	if err != nil {
		return nil, err
	}
	w.Close()
	return buf.Bytes(), nil
}

// load key and cert from datastore
func loadRootCA() error {
	if datastore.PKIConf.RootCAKey == "" {
		log.Printf("no ca")
		return nil
	}
	key, pub, err := getPrivateKeyFromPEM(datastore.PKIConf.RootCAKey)
	if err != nil {
		return err
	}
	rootCAPrivateKey = key
	rootCAPublicKey = pub
	cert, err := getCertFromPEM(datastore.PKIConf.RootCACert)
	if err != nil {
		return err
	}
	rootCACertificate = cert
	return nil
}

func getPrivateKeyFromPEM(p string) (any, any, error) {
	block, _ := pem.Decode([]byte(p))
	if block == nil {
		return nil, nil, fmt.Errorf("invalid private key data")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, nil, err
		}
		return key, &key.PublicKey, nil
	case "EC PRIVATE KEY":
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, nil, err
		}
		return key, &key.PublicKey, nil
	default:
		return nil, nil, fmt.Errorf("unknown keytype type=%s", block.Type)
	}
}

func getCertFromPEM(p string) ([]byte, error) {
	block, _ := pem.Decode([]byte(p))
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("invalid cert pem data")
	}
	return block.Bytes, nil
}

func makePEM(data []byte, pemType string) []byte {
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  pemType,
			Bytes: data,
		},
	)
}

func createRootCACertificate() error {
	var keyBytes []byte
	var err error
	var isRSA = false
	if strings.HasPrefix(datastore.PKIConf.RootCAKeyType, "rsa-") {
		bits := 4096
		switch datastore.PKIConf.RootCAKeyType {
		case "rsa-2048":
			bits = 2048
		case "rsa-8192":
			bits = 8192
		}
		key, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			return err
		}
		rootCAPrivateKey = key
		rootCAPublicKey = &key.PublicKey
		keyBytes = x509.MarshalPKCS1PrivateKey(key)
		isRSA = true
	} else {
		var curve = elliptic.P256()
		switch datastore.PKIConf.RootCAKeyType {
		case "ecdsa-224":
			curve = elliptic.P224()
		case "ecdsa-384":
			curve = elliptic.P384()
		case "ecdsa-521":
			curve = elliptic.P521()
		}
		key, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			return err
		}
		rootCAPrivateKey = key
		rootCAPublicKey = &key.PublicKey
		keyBytes, err = x509.MarshalECPrivateKey(key)
		if err != nil {
			return err
		}
	}
	subject := pkix.Name{
		CommonName:   datastore.PKIConf.Name + " Root CA",
		Organization: []string{datastore.PKIConf.Name},
	}
	sn := getSerial()
	tmp := &x509.Certificate{
		SerialNumber:          big.NewInt(sn),
		Subject:               subject,
		NotAfter:              time.Now().AddDate(datastore.PKIConf.RootCATerm, 0, 0).UTC(),
		NotBefore:             time.Now().UTC(),
		IsCA:                  true,
		MaxPathLen:            int(1),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	rootCACertificate, err = x509.CreateCertificate(rand.Reader, tmp, tmp, rootCAPublicKey, rootCAPrivateKey)
	if err != nil {
		return err
	}
	// 証明書を発行
	datastore.PKIConf.RootCACert = string(makePEM(rootCACertificate, "CERTIFICATE"))
	if isRSA {
		datastore.PKIConf.RootCAKey = string(makePEM(keyBytes, "RSA PRIVATE KEY"))
	} else {
		datastore.PKIConf.RootCAKey = string(makePEM(keyBytes, "EC PRIVATE KEY"))
	}
	datastore.UpdateCert(&datastore.PKICertEnt{
		ID:          fmt.Sprintf("%x", sn),
		Subject:     subject.String(),
		Created:     time.Now().UnixNano(),
		Certificate: datastore.PKIConf.RootCACert,
		Expire:      tmp.NotAfter.UnixNano(),
		Type:        "system",
	})
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:  time.Now().UnixNano(),
		Type:  "ca",
		Level: "info",
		Event: fmt.Sprintf("CA証明書を発行しました subject=%s serial=%x", subject.String(), sn),
	})
	return nil
}

// シリアル番号の取得は排他的に行う
var snMu sync.Mutex

func getSerial() int64 {
	snMu.Lock()
	defer snMu.Unlock()
	sn := datastore.PKIConf.Serial
	datastore.PKIConf.Serial++
	log.Printf("update ca serial %d", datastore.PKIConf.Serial)
	return sn
}

// CreateCertificate manually issues a certificate from a given CSR.
func CreateCertificate(csr []byte) ([]byte, error) {
	block, _ := pem.Decode(csr)
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		return nil, fmt.Errorf("CSR not found")
	}
	crt, _, err := createCertificateFromCSR(block.Bytes, "manual", make(map[string]string))
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: crt,
		},
	), nil
}

// createCertificateFromCSR : CSRから証明書を発行する
func createCertificateFromCSR(csrBytes []byte, certType string, info map[string]string) ([]byte, string, error) {
	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		return nil, "", err
	}
	nodeID, err := checkCSR(certType, csr, info)
	if err != nil {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Time:  time.Now().UnixNano(),
			Type:  "ca",
			Level: "low",
			Event: fmt.Sprintf("証明書の発行を却下しました subject=%s info=%+v err=%v", csr.Subject.String(), info, err),
		})
		return nil, "", err
	}
	ca, err := x509.ParseCertificate(rootCACertificate)
	if err != nil {
		return nil, "", err
	}
	term := datastore.PKIConf.CertTerm
	if term < 1 {
		term = 24 * 30
	}
	sn := getSerial()
	tmp := &x509.Certificate{
		SerialNumber:          big.NewInt(sn),
		Subject:               csr.Subject,
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().Add(time.Hour * time.Duration(term)).UTC(),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		DNSNames:              csr.DNSNames,
		EmailAddresses:        csr.EmailAddresses,
		IPAddresses:           csr.IPAddresses,
		ExtraExtensions:       csr.Extensions,
		IsCA:                  false,
		MaxPathLen:            0,
		BasicConstraintsValid: true,
		CRLDistributionPoints: []string{},
		OCSPServer:            []string{},
	}
	for _, san := range strings.Split(datastore.PKIConf.SANs, ",") {
		baseURL := fmt.Sprintf("http://%s:%d/", san, datastore.PKIConf.HTTPPort)
		tmp.CRLDistributionPoints = append(tmp.CRLDistributionPoints, baseURL+"crl")
		tmp.OCSPServer = append(tmp.OCSPServer, baseURL+"ocsp")
	}
	if strings.HasPrefix(datastore.PKIConf.HTTPBaseURL, "http://") {
		baseURL := strings.TrimRight(datastore.PKIConf.HTTPBaseURL, "/")
		tmp.CRLDistributionPoints = append(tmp.CRLDistributionPoints, baseURL+"/crl")
		tmp.OCSPServer = append(tmp.OCSPServer, baseURL+"/ocsp")
	}
	ret, err := x509.CreateCertificate(rand.Reader, tmp, ca, csr.PublicKey, rootCAPrivateKey)
	if err != nil {
		return nil, "", err
	}
	certID := fmt.Sprintf("%x", sn)
	datastore.UpdateCert(&datastore.PKICertEnt{
		ID:          certID,
		Subject:     tmp.Subject.String(),
		Created:     time.Now().UnixNano(),
		Expire:      tmp.NotAfter.UnixNano(),
		NodeID:      nodeID,
		Info:        info,
		Certificate: string(makePEM(ret, "CERTIFICATE")),
		Type:        certType,
	})
	nodeName := ""
	if node := datastore.GetNode(nodeID); node != nil {
		nodeName = node.Name
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:     time.Now().UnixNano(),
		Type:     "ca",
		Level:    "info",
		NodeID:   nodeID,
		NodeName: nodeName,
		Event:    fmt.Sprintf("証明書を発行しました subject=%s serial=%s info=%+v", tmp.Subject.String(), certID, info),
	})
	return ret, certID, nil
}

func checkCSR(certType string, csr *x509.CertificateRequest, info map[string]string) (string, error) {
	var node *datastore.NodeEnt
	for _, n := range csr.DNSNames {
		if node = datastore.FindNodeFromName(n); node != nil {
			break
		}
	}
	if node == nil {
		for _, ip := range csr.IPAddresses {
			if ipv4 := ip.To4(); ipv4 != nil {
				if node = datastore.FindNodeFromIP(ipv4.String()); node != nil {
					break
				}
			}
		}
		if node == nil {
			node = datastore.FindNodeFromName(csr.Subject.CommonName)
			if node == nil {
				if ip, ok := info["RemoteAddr"]; ok {
					node = datastore.FindNodeFromIP(ip)
				}
			}
		}
	}
	if node == nil {
		return "", fmt.Errorf("node not found")
	}
	if certType == "scep" {
		if p, ok := info["ChallengePassword"]; !ok || p == "" || node.Password != p {
			return "", fmt.Errorf("challenge password mismatch")
		}
	}
	return node.ID, nil
}

// CRL
// RFC5280, 5.2.5
type issuingDistributionPoint struct {
	DistributionPoint          distributionPointName `asn1:"optional,tag:0"`
	OnlyContainsUserCerts      bool                  `asn1:"optional,tag:1"`
	OnlyContainsCACerts        bool                  `asn1:"optional,tag:2"`
	OnlySomeReasons            asn1.BitString        `asn1:"optional,tag:3"`
	IndirectCRL                bool                  `asn1:"optional,tag:4"`
	OnlyContainsAttributeCerts bool                  `asn1:"optional,tag:5"`
}

type distributionPointName struct {
	FullName     []asn1.RawValue  `asn1:"optional,tag:0"`
	RelativeName pkix.RDNSequence `asn1:"optional,tag:1"`
}

func createCRL() error {
	dp := distributionPointName{
		FullName: []asn1.RawValue{},
	}
	for _, san := range strings.Split(datastore.PKIConf.SANs, ",") {
		cdp := fmt.Sprintf("http://%s:%d/crl", san, datastore.PKIConf.HTTPPort)
		dp.FullName = append(dp.FullName, asn1.RawValue{Tag: 6, Class: 2, Bytes: []byte(cdp)})
	}
	var oidExtensionIssuingDistributionPoint = []int{2, 5, 29, 28}
	idp := issuingDistributionPoint{
		DistributionPoint: dp,
	}
	v, err := asn1.Marshal(idp)
	if err != nil {
		return err
	}

	cdpExt := pkix.Extension{
		Id:       oidExtensionIssuingDistributionPoint,
		Critical: true,
		Value:    v,
	}

	key, ok := rootCAPrivateKey.(crypto.Signer)
	if !ok {
		return fmt.Errorf("invalid key type")
	}
	ca, err := x509.ParseCertificate(rootCACertificate)
	if err != nil {
		return err
	}
	revokedCerts := []pkix.RevokedCertificate{}
	now := time.Now().UnixNano()
	datastore.ForEachCert(func(c *datastore.PKICertEnt) bool {
		// たぶん、期限切れは含めない
		if c.Revoked > 0 && c.Expire > now {
			if s, ok := big.NewInt(0).SetString(c.ID, 16); ok {
				revokedCerts = append(revokedCerts, pkix.RevokedCertificate{
					RevocationTime: time.Unix(0, c.Revoked),
					SerialNumber:   s,
				})
			}
		}
		return true
	})
	tmp := &x509.RevocationList{
		SignatureAlgorithm:  x509.ECDSAWithSHA256,
		RevokedCertificates: revokedCerts,
		Number:              big.NewInt(datastore.PKIConf.CrlNumber),
		ThisUpdate:          time.Now(),
		NextUpdate:          time.Now().Add(time.Duration(datastore.PKIConf.CrlInterval) * time.Hour),
		ExtraExtensions:     []pkix.Extension{cdpExt},
	}
	crl, err = x509.CreateRevocationList(rand.Reader, tmp, ca, key)
	if err == nil {
		datastore.PKIConf.CrlNumber++
	}
	return err
}
