package webapi

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type mibbrWebAPI struct {
	Node    *datastore.NodeEnt
	MIBTree *[]*datastore.MIBTreeEnt
}

func getMIBBr(c echo.Context) error {
	id := c.Param("id")
	r := mibbrWebAPI{}
	if strings.HasPrefix(id, "NET:") {
		nt := datastore.GetNetwork(id)
		if nt != nil {
			r.Node = &datastore.NodeEnt{
				ID:   id,
				Name: nt.Name,
				IP:   nt.IP,
			}
		}
	} else {
		r.Node = datastore.GetNode(id)
	}
	r.MIBTree = &datastore.MIBTree
	if r.Node == nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, r)
}

type mibGetReqWebAPI struct {
	NodeID string
	Name   string
	OID    string
	Raw    bool
}

type mibEnt struct {
	Name  string
	Value string
}

func postMIBBr(c echo.Context) error {
	p := new(mibGetReqWebAPI)
	if err := c.Bind(p); err != nil {
		log.Printf("MIBBR err=%v", err)
		return echo.ErrBadRequest
	}
	r, err := snmpWalk(p)
	if err != nil {
		log.Printf("MIBBR err=%v", err)
		if len(r) > 0 {
			return c.JSON(http.StatusContinue, r)
		}
		return echo.ErrBadRequest
	}
	if len(r) < 1 {
		log.Println("MIBBR not found")
		return echo.ErrNotFound
	}
	return c.JSON(http.StatusOK, r)
}

type mibSetReqWebAPI struct {
	NodeID string
	Name   string
	Type   string
	Value  string
}

func postSNMPSet(c echo.Context) error {
	p := new(mibSetReqWebAPI)
	if err := c.Bind(p); err != nil {
		log.Printf("SNMP set err=%v", err)
		return echo.ErrBadRequest
	}
	r := snmpset(p)
	return c.String(http.StatusOK, r)
}

func nameToOID(name string) string {
	oid := datastore.MIBDB.NameToOID(name)
	if oid == ".1" {
		oid = ".1.3"
	}
	if oid == ".0.0" {
		if matched, _ := regexp.MatchString(`\.[0-9.]+`, name); matched {
			return name
		}
	}
	if !strings.HasPrefix(oid, ".") {
		oid = "." + oid
	}
	return oid
}

func snmpWalk(p *mibGetReqWebAPI) ([]*mibEnt, error) {
	ret := []*mibEnt{}
	agent, err := getSNMPAgent(p.NodeID)
	if err != nil {
		return ret, err
	}
	if err := agent.Connect(); err != nil {
		return ret, err
	}
	et := time.Now().Unix() + (3 * 60)
	defer agent.Conn.Close()
	err = agent.Walk(nameToOID(p.Name), func(variable gosnmp.SnmpPDU) error {
		if et < time.Now().Unix() {
			return fmt.Errorf("timeout")
		}
		name := datastore.MIBDB.OIDToName(variable.Name)
		value := datastore.GetMIBValueString(name, &variable, p.Raw)
		ret = append(ret, &mibEnt{
			Name:  name,
			Value: value,
		})
		return nil
	})
	if err != nil {
		log.Printf("n=%s o=%s err=%v", p.Name, nameToOID(p.Name), err)
	}
	return ret, err
}

func snmpset(p *mibSetReqWebAPI) string {
	agent, err := getSNMPAgent(p.NodeID)
	if err != nil {
		return err.Error()
	}
	if err := agent.Connect(); err != nil {
		return err.Error()
	}
	defer agent.Conn.Close()
	setPDU := []gosnmp.SnmpPDU{}
	switch p.Type {
	case "integer":
		i, err := strconv.Atoi(p.Value)
		if err != nil {
			return err.Error()
		}
		setPDU = append(setPDU, gosnmp.SnmpPDU{
			Name:  nameToOID(p.Name),
			Type:  gosnmp.Integer,
			Value: i,
		})
	default:
		// string
		setPDU = append(setPDU, gosnmp.SnmpPDU{
			Name:  nameToOID(p.Name),
			Type:  gosnmp.OctetString,
			Value: []byte(p.Value),
		})
	}
	r, err := agent.Set(setPDU)
	if err != nil {
		return err.Error()
	}
	if r.Error != gosnmp.NoError {
		return r.Error.String()
	}
	return ""
}

func getSNMPAgent(nodeID string) (*gosnmp.GoSNMP, error) {
	n := datastore.GetNode(nodeID)
	if n == nil {
		if !strings.HasPrefix(nodeID, "NET:") {
			return nil, fmt.Errorf("node not found")
		}
		nt := datastore.GetNetwork(nodeID)
		if nt == nil {
			return nil, fmt.Errorf("network not found")
		}
		n = &datastore.NodeEnt{
			IP:        nt.IP,
			SnmpMode:  nt.SnmpMode,
			Community: nt.Community,
			User:      nt.User,
			Password:  nt.Password,
		}
	}
	agent := &gosnmp.GoSNMP{
		Target:    n.IP,
		Port:      161,
		Transport: "udp",
		Community: n.Community,
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(datastore.MapConf.Timeout) * time.Second,
		Retries:   datastore.MapConf.Retry,
		MaxOids:   gosnmp.MaxOids,
	}
	switch n.SnmpMode {
	case "v3auth":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthNoPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: n.Password,
		}
	case "v3authpriv":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: n.Password,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        n.Password,
		}
	case "v3authprivex":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA256,
			AuthenticationPassphrase: n.Password,
			PrivacyProtocol:          gosnmp.AES256,
			PrivacyPassphrase:        n.Password,
		}
	}
	return agent, nil
}
