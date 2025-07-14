package pki

import (
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/smallstep/scep"

	"github.com/twsnmp/twsnmpfc/datastore"
)

var scepCAPrivateKey any
var scepCAPublicKey any
var scepCACertificate []byte

func createScepCACertificate() error {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	scepCAPrivateKey = key
	scepCAPublicKey = &key.PublicKey
	ca, err := x509.ParseCertificate(rootCACertificate)
	if err != nil {
		return err
	}
	sn := getSerial()
	tmp := &x509.Certificate{
		SerialNumber: big.NewInt(sn),
		Subject: pkix.Name{
			CommonName: datastore.PKIConf.Name + " SCEP CA",
		},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().AddDate(datastore.PKIConf.RootCATerm, 0, 0).UTC(),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
		MaxPathLen:            int(1),
	}
	for _, san := range strings.Split(datastore.PKIConf.SANs, ",") {
		baseURL := fmt.Sprintf("http://%s:%d", san, datastore.PKIConf.HTTPPort)
		if ip := net.ParseIP(san); ip == nil {
			tmp.DNSNames = append(tmp.DNSNames, san)
		} else {
			tmp.IPAddresses = append(tmp.IPAddresses, ip)
		}
		tmp.CRLDistributionPoints = append(tmp.CRLDistributionPoints, baseURL+"/crl")
		tmp.OCSPServer = append(tmp.OCSPServer, baseURL+"/ocsp")
	}
	if strings.HasPrefix(datastore.PKIConf.HTTPBaseURL, "http://") {
		baseURL := strings.TrimRight(datastore.PKIConf.HTTPBaseURL, "/")
		tmp.CRLDistributionPoints = append(tmp.CRLDistributionPoints, baseURL+"/crl")
		tmp.OCSPServer = append(tmp.OCSPServer, baseURL+"/ocsp")
	}
	scepCACertificate, err = x509.CreateCertificate(rand.Reader, tmp, ca, scepCAPublicKey, rootCAPrivateKey)
	if err != nil {
		return err
	}
	datastore.PKIConf.ScepCACert = string(makePEM(scepCACertificate, "CERTIFICATE"))
	datastore.PKIConf.ScepCAKey = string(makePEM(x509.MarshalPKCS1PrivateKey(key), "RSA PRIVATE KEY"))
	datastore.UpdateCert(&datastore.PKICertEnt{
		ID:          fmt.Sprintf("%x", sn),
		Subject:     tmp.Subject.String(),
		Created:     time.Now().UnixNano(),
		Expire:      tmp.NotAfter.UnixNano(),
		Certificate: datastore.PKIConf.ScepCACert,
		Type:        "system",
	})
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:  time.Now().UnixNano(),
		Type:  "ca",
		Level: "info",
		Event: fmt.Sprintf("SCEP用CAの証明書を発行しました subject=%s serial=%x", tmp.Subject.String(), sn),
	})
	return nil
}

func loadScepCA() error {
	if datastore.PKIConf.ScepCAKey == "" {
		log.Printf("no scep ca")
		return nil
	}
	key, pub, err := getPrivateKeyFromPEM(datastore.PKIConf.ScepCAKey)
	if err != nil {
		return err
	}
	scepCAPrivateKey = key
	scepCAPublicKey = pub
	cert, err := getCertFromPEM(datastore.PKIConf.ScepCACert)
	if err != nil {
		return err
	}
	scepCACertificate = cert
	return nil
}

const maxPayloadSize = 2 << 20

// request is a SCEP server request.
type request struct {
	Operation string
	Message   []byte
}

func scepFunc(c echo.Context) error {
	req, err := decodeRequest(c)
	if err != nil {
		log.Printf("scep err=%v", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	switch req.Operation {
	case "GetCACert":
		return getCACert(c)
	case "GetCACaps":
		return getCACaps(c)
	case "PKIOperation":
		return pkiOperation(c, req)
	default:
		log.Printf("scep unknown operation op=%v", req.Operation)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("unknown operation: %s", req.Operation))
	}
}

func decodeRequest(c echo.Context) (request, error) {
	defer c.Request().Body.Close()
	operation := c.QueryParam("operation")
	var err error
	req := request{
		Operation: operation,
	}
	if c.Request().Method == http.MethodPost {
		req.Message, err = io.ReadAll(io.LimitReader(c.Request().Body, maxPayloadSize))
		if err != nil {
			return request{}, fmt.Errorf("failed reading request body: %w", err)
		}
	} else {
		message := c.QueryParam("message")
		req.Message, err = decodeMessage(message)
		if err != nil {
			return request{}, fmt.Errorf("failed decoding message: %w", err)
		}
	}
	return req, nil
}

func decodeMessage(message string) ([]byte, error) {
	if message == "" {
		return nil, nil
	}
	decodedMessage, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return nil, err
	}
	return decodedMessage, nil
}

// getCACert returns the CA certificates in a SCEP response
func getCACert(c echo.Context) error {
	certs := []*x509.Certificate{}
	ca, err := x509.ParseCertificate(rootCACertificate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	certs = append(certs, ca)
	scepCA, err := x509.ParseCertificate(scepCACertificate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	certs = append(certs, scepCA)
	if len(certs) == 1 {
		return c.Blob(http.StatusOK, "application/x-x509-ca-cert", ca.Raw)
	}
	data, err := scep.DegenerateCertificates(certs)
	if err != nil {
		log.Printf(" err=%v", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.Blob(http.StatusOK, "application/x-x509-ca-ra-cert", data)
}

// getCACaps returns the CA capabilities in a SCEP response
func getCACaps(c echo.Context) error {
	caps := []string{
		"Renewal",
		"SHA-1",
		"SHA-256",
		"AES",
		"DES3",
		"SCEPStandard",
		"POSTPKIOperation",
	}
	return c.String(http.StatusOK, strings.Join(caps, "\r\n"))
}

// pkiOperation performs PKI operations and returns a SCEP response
func pkiOperation(c echo.Context, req request) error {
	msg, err := scep.ParsePKIMessage(req.Message)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	// key, ok := rootCAPrivateKey.(crypto.Signer)
	key, ok := scepCAPrivateKey.(crypto.Signer)
	if !ok {
		log.Printf("err=%v", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	// ca, err := x509.ParseCertificate(rootCACertificate)
	ca, err := x509.ParseCertificate(scepCACertificate)
	if err != nil {
		log.Printf("err=%v", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	// extract encrypted pkiEnvelope
	err = msg.DecryptPKIEnvelope(ca, key)
	if err != nil {
		log.Printf("err=%v", err)
		return c.JSON(http.StatusBadRequest, err)
	}
	csr := msg.CSR
	transactionID := string(msg.TransactionID)
	challengePassword := msg.ChallengePassword
	log.Printf("transactionID=%v  challengePassword=%v fromIP=%s", transactionID, challengePassword, c.RealIP())
	crtBytes, _, err := createCertificateFromCSR(csr.Raw, "scep",
		map[string]string{"RemoteAddr": c.RealIP(), "TransactionID": transactionID, "ChallengePassword": challengePassword})
	if err != nil {
		log.Printf("create cert err=%v", err)
		rsp, err := msg.Fail(ca, key, scep.BadRequest)
		if err != nil {
			log.Printf("err=%v", err)
			return c.JSON(http.StatusInternalServerError, err)
		}
		return c.Blob(http.StatusOK, "application/x-pki-message", rsp.Raw)
	}
	crt, err := x509.ParseCertificate(crtBytes)
	if err != nil {
		log.Printf("err=%v", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	rsp, err := msg.Success(ca, key, crt)
	if err != nil {
		log.Printf("err=%v", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.Blob(http.StatusOK, "application/x-pki-message", rsp.Raw)
}
