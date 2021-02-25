package webapi

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/gosnmp"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type mibbrWebAPI struct {
	Node    *datastore.NodeEnt
	MIBTree *[]*datastore.MIBTreeEnt
}

func getMIBBr(c echo.Context) error {
	id := c.Param("id")
	api := c.Get("api").(*WebAPI)
	r := mibbrWebAPI{}
	r.Node = api.DataStore.GetNode(id)
	r.MIBTree = &datastore.MIBTree
	if r.Node == nil {
		log.Printf("node not found")
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, r)
}

type mibGetReqWebAPI struct {
	NodeID string
	Name   string
	OID    string
}

type mibEnt struct {
	Name  string
	Value string
}

func postMIBBr(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	p := new(mibGetReqWebAPI)
	if err := c.Bind(p); err != nil {
		log.Printf("postSyslog err=%v", err)
		return echo.ErrBadRequest
	}
	r, err := snmpWalk(api, p)
	if err != nil {
		log.Printf("postMIBBr err=%v", err)
		return echo.ErrBadRequest
	}
	if len(r) < 1 {
		log.Printf("postMIBBr no mibs")
		return echo.ErrNotFound
	}
	return c.JSON(http.StatusOK, r)
}

func snmpWalk(api *WebAPI, p *mibGetReqWebAPI) ([]*mibEnt, error) {
	ret := []*mibEnt{}
	n := api.DataStore.GetNode(p.NodeID)
	if n == nil {
		log.Printf("postMIBBr node not found")
		return ret, fmt.Errorf("node not found")
	}
	agent := &gosnmp.GoSNMP{
		Target:             n.IP,
		Port:               161,
		Transport:          "udp",
		Community:          n.Community,
		Version:            gosnmp.Version2c,
		Timeout:            time.Duration(api.DataStore.MapConf.Timeout) * time.Second,
		Retries:            api.DataStore.MapConf.Retry,
		ExponentialTimeout: true,
		MaxOids:            gosnmp.MaxOids,
	}
	if n.SnmpMode != "" {
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		if n.SnmpMode == "v3auth" {
			agent.MsgFlags = gosnmp.AuthNoPriv
			agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 n.User,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: n.Password,
			}
		} else {
			agent.MsgFlags = gosnmp.AuthPriv
			agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 n.User,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: n.Password,
				PrivacyProtocol:          gosnmp.AES,
				PrivacyPassphrase:        n.Password,
			}
		}
	}
	err := agent.Connect()
	if err != nil {
		return ret, err
	}
	defer agent.Conn.Close()
	err = agent.Walk(api.DataStore.MIBDB.NameToOID(p.Name), func(variable gosnmp.SnmpPDU) error {
		name := api.DataStore.MIBDB.OIDToName(variable.Name)
		value := ""
		if variable.Type == gosnmp.OctetString {
			if strings.Contains(api.DataStore.MIBDB.OIDToName(variable.Name), "ifPhysAd") {
				a := variable.Value.(string)
				if len(a) > 5 {
					value = fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", a[0], a[1], a[2], a[3], a[4], a[5])
				}
			} else {
				value = variable.Value.(string)
			}
		} else if variable.Type == gosnmp.ObjectIdentifier {
			value = api.DataStore.MIBDB.OIDToName(variable.Value.(string))
		} else {
			value = fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Uint64())
		}
		ret = append(ret, &mibEnt{
			Name:  name,
			Value: value,
		})
		return nil
	})
	return ret, err
}
