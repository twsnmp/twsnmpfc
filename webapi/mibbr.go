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
	r := mibbrWebAPI{}
	r.Node = datastore.GetNode(id)
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
}

type mibEnt struct {
	Name  string
	Value string
}

func postMIBBr(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	p := new(mibGetReqWebAPI)
	if err := c.Bind(p); err != nil {
		return echo.ErrBadRequest
	}
	r, err := snmpWalk(api, p)
	if err != nil {
		return echo.ErrBadRequest
	}
	if len(r) < 1 {
		return echo.ErrNotFound
	}
	return c.JSON(http.StatusOK, r)
}

func snmpWalk(api *WebAPI, p *mibGetReqWebAPI) ([]*mibEnt, error) {
	ret := []*mibEnt{}
	n := datastore.GetNode(p.NodeID)
	if n == nil {
		return ret, fmt.Errorf("node not found")
	}
	agent := &gosnmp.GoSNMP{
		Target:             n.IP,
		Port:               161,
		Transport:          "udp",
		Community:          n.Community,
		Version:            gosnmp.Version2c,
		Timeout:            time.Duration(datastore.MapConf.Timeout) * time.Second,
		Retries:            datastore.MapConf.Retry,
		ExponentialTimeout: true,
		MaxOids:            gosnmp.MaxOids,
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
	err := agent.Connect()
	if err != nil {
		return ret, err
	}
	defer agent.Conn.Close()
	err = agent.Walk(datastore.MIBDB.NameToOID(p.Name), func(variable gosnmp.SnmpPDU) error {
		name := datastore.MIBDB.OIDToName(variable.Name)
		value := ""
		switch variable.Type {
		case gosnmp.OctetString:
			mi := datastore.FindMIBInfo(name)
			if mi != nil {
				switch mi.Type {
				case "PhysAddress", "OctetString":
					a, ok := variable.Value.([]uint8)
					if !ok {
						a = []uint8(getMIBStringVal(variable.Value))
					}
					mac := []string{}
					for _, m := range a {
						mac = append(mac, fmt.Sprintf("%02X", m&0x00ff))
					}
					value = strings.Join(mac, ":")
				case "DisplayString":
					value = getMIBStringVal(variable.Value)
				default:
					log.Printf("%s=%s:%v:%v", name, mi.Type, variable.Type, variable.Value)
					value = getMIBStringVal(variable.Value)
				}
			} else {
				value = getMIBStringVal(variable.Value)
			}
		case gosnmp.ObjectIdentifier:
			value = datastore.MIBDB.OIDToName(getMIBStringVal(variable.Value))
		case gosnmp.TimeTicks:
			t := gosnmp.ToBigInt(variable.Value).Uint64()
			if t > (24 * 3600 * 100) {
				d := t / (24 * 3600 * 100)
				t -= d * (24 * 3600 * 100)
				value = fmt.Sprintf("%d(%d days, %v)", t, d, time.Duration(t*10*uint64(time.Millisecond)))
			} else {
				value = fmt.Sprintf("%d(%v)", t, time.Duration(t*10*uint64(time.Millisecond)))
			}
		default:
			v := int(gosnmp.ToBigInt(variable.Value).Uint64())
			vname := ""
			mi := datastore.FindMIBInfo(name)
			if mi != nil && mi.Enum != "" {
				if vn, ok := mi.EnumMap[v]; ok {
					vname = "(" + vn + ")"
				}
			}
			value = fmt.Sprintf("%d%s", gosnmp.ToBigInt(variable.Value).Uint64(), vname)
		}
		ret = append(ret, &mibEnt{
			Name:  name,
			Value: value,
		})
		return nil
	})
	return ret, err
}

func getMIBStringVal(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	case []uint8:
		return string(v)
	case int, int64, uint, uint64:
		return fmt.Sprintf("%d", v)
	}
	return ""
}
