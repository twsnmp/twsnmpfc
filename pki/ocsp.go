package pki

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"golang.org/x/crypto/ocsp"
)

func ocspFunc(c echo.Context, b []byte) error {
	req, err := ocsp.ParseRequest(b)
	if err != nil {
		log.Printf("err=%v", err)
		return c.JSON(http.StatusBadRequest, err)
	}
	key, ok := rootCAPrivateKey.(crypto.Signer)
	if !ok {
		log.Printf("err=%v", err)
		return c.JSON(http.StatusBadRequest, err)
	}
	ca, err := x509.ParseCertificate(rootCACertificate)
	if err != nil {
		log.Printf("err=%v", err)
		return c.JSON(http.StatusBadRequest, err)
	}
	res := ocsp.Response{
		Status:       ocsp.Unknown,
		SerialNumber: req.SerialNumber,
		ThisUpdate:   time.Now(),
		NextUpdate:   time.Now().Add(time.Second * 60 * 10),
		IssuerHash:   req.HashAlgorithm,
	}
	if bytes.Equal(req.IssuerKeyHash, ca.SubjectKeyId) {
		id := fmt.Sprintf("%x", req.SerialNumber.Int64())
		cert := datastore.FindCert(id)
		if cert.Revoked != 0 {
			res.Status = ocsp.Revoked
			res.RevokedAt = time.Now().Add(time.Hour * -3)
			res.RevocationReason = ocsp.Unspecified
		} else {
			res.Status = ocsp.Good
		}
	}
	bres, err := ocsp.CreateResponse(ca, ca, res, key)
	if err != nil {
		log.Printf("err=%v", err)
		return c.JSON(http.StatusBadRequest, err)
	}
	return c.Blob(http.StatusOK, "application/ocsp-response", bres)
}
