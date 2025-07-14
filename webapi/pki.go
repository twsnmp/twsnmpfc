package webapi

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/pki"
)

// getHasCA: CAが構築済みかを返す
func getHasCA(c echo.Context) error {
	return c.JSON(http.StatusOK, pki.IsCAValid())
}

// CA作成リクエストのデフォルト値を取得する
func getDefaultCreateCAReq(c echo.Context) error {
	return c.JSON(http.StatusOK, &datastore.CreateCAReq{
		RootCAKeyType: datastore.PKIConf.RootCAKeyType,
		Name:          datastore.PKIConf.Name,
		SANs:          datastore.PKIConf.SANs,
		AcmeBaseURL:   datastore.PKIConf.AcmeBaseURL,
		AcmePort:      datastore.PKIConf.AcmePort,
		HTTPBaseURL:   datastore.PKIConf.HTTPBaseURL,
		HTTPPort:      datastore.PKIConf.HTTPPort,
		RootCATerm:    datastore.PKIConf.RootCATerm,
		CrlInterval:   datastore.PKIConf.CrlInterval,
		CertTerm:      datastore.PKIConf.CertTerm,
	})
}

// CAを作成する
func postCreateCA(c echo.Context) error {
	if pki.IsCAValid() {
		return echo.ErrBadRequest
	}
	req := new(datastore.CreateCAReq)
	if err := c.Bind(req); err != nil {
		log.Printf("postCreateCA err=%v", err)
		return echo.ErrBadRequest
	}
	if err := pki.CreateCA(req); err != nil {
		log.Printf("postCreateCA err=%v", err)
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:  time.Now().UnixNano(),
		Type:  "ca",
		Level: "info",
		Event: "CAを構築しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postDestroyCA(c echo.Context) error {
	pki.DestroyCA()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:  time.Now().UnixNano(),
		Type:  "ca",
		Level: "info",
		Event: "CAを破棄しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

type CertEnt struct {
	Status  string `json:"Status"`
	ID      string `json:"ID"`
	Subject string `json:"Subject"`
	Node    string `json:"Node"`
	Created int64  `json:"Created"`
	Revoked int64  `json:"Revoked"`
	Expire  int64  `json:"Expire"`
	Type    string `json:"Type"`
}

// getPKICerts: 証明書のリストを返す
func getPKICerts(c echo.Context) error {
	ret := []*CertEnt{}
	now := time.Now().UnixNano()
	datastore.ForEachCert(func(c *datastore.PKICertEnt) bool {
		status := "valid"
		if c.Revoked > 0 {
			status = "revoked"
		} else if c.Expire < now {
			status = "expired"
		}
		node := ""
		if c.NodeID != "" {
			if n := datastore.GetNode(c.NodeID); n != nil {
				node = n.Name
			}
		}
		ret = append(ret, &CertEnt{
			Status:  status,
			ID:      c.ID,
			Subject: c.Subject,
			Node:    node,
			Created: c.Created,
			Revoked: c.Revoked,
			Expire:  c.Expire,
			Type:    c.Type,
		})
		return true
	})
	return c.JSON(http.StatusOK, &ret)
}

func postCreateCertificateRequest(c echo.Context) error {
	req := new(pki.CSRReqEnt)
	err := c.Bind(req)
	if err != nil {
		log.Printf("postCreateCertificateRequest err=%v", err)
		return echo.ErrBadRequest
	}
	b, err := pki.CreateCertificateRequest(req)
	if err != nil {
		log.Printf("postCreateCertificateRequest err=%v", err)
		return echo.ErrBadRequest
	}
	return c.Blob(http.StatusOK, "application/zip", b)
}

func postCreateCertificate(c echo.Context) error {
	f, err := c.FormFile("file")
	if err != nil {
		return echo.ErrBadRequest
	}
	if f.Size > 1024*1024*10 {
		return echo.ErrBadRequest
	}
	file, err := f.Open()
	if err != nil {
		return echo.ErrBadRequest
	}
	defer file.Close()
	b := make([]byte, f.Size)
	_, err = file.Read(b)
	if err != nil {
		return echo.ErrBadRequest
	}
	crt, err := pki.CreateCertificate(b)
	if err != nil {
		log.Printf("create Certificate err=%v", err)
		return echo.ErrBadRequest
	}
	return c.Blob(http.StatusOK, "application/x-pem-file", crt)
}

// deleteRevokeCert: 証明書の失効
func deleteRevokeCert(c echo.Context) error {
	id := c.Param("id")
	datastore.RevokeCertByID(id)
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

// getExportCert: 証明書をエクスポート
func getExportCert(c echo.Context) error {
	id := c.Param("id")
	cert := datastore.FindCert(id)
	if cert == nil {
		return echo.ErrNotFound
	}
	return c.Blob(http.StatusOK, "application/x-pem-file", []byte(cert.Certificate))
}

func getPKIControl(c echo.Context) error {
	return c.JSON(http.StatusOK, &datastore.PKIControlEnt{
		EnableAcme:  datastore.PKIConf.EnableAcme,
		EnableHTTP:  datastore.PKIConf.EnableHTTP,
		AcmeBaseURL: datastore.PKIConf.AcmeBaseURL,
		CertTerm:    datastore.PKIConf.CertTerm,
		CrlInterval: datastore.PKIConf.CrlInterval,
		AcmeStatus:  pki.GetAcmeServerStatus(),
		HTTPStatus:  pki.GetHTTPServerStatus(),
	})
}

func postPKIControl(c echo.Context) error {
	req := new(datastore.PKIControlEnt)
	if err := c.Bind(req); err != nil {
		log.Printf("create Certificate err=%v", err)
		return echo.ErrBadRequest
	}
	datastore.PKIConf.EnableAcme = req.EnableAcme
	datastore.PKIConf.EnableHTTP = req.EnableHTTP
	datastore.PKIConf.AcmeBaseURL = req.AcmeBaseURL
	datastore.PKIConf.CertTerm = req.CertTerm
	datastore.PKIConf.CrlInterval = req.CrlInterval
	datastore.SavePKIConf()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
