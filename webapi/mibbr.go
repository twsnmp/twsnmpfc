package webapi

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/logger"
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
	Raw    bool
}

type mibEnt struct {
	Name  string
	Value string
}

func postMIBBr(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	p := new(mibGetReqWebAPI)
	if err := c.Bind(p); err != nil {
		log.Printf("MIBBR err=%v", err)
		return echo.ErrBadRequest
	}
	r, err := snmpWalk(api, p)
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

func nameToOID(name string) string {
	oid := datastore.MIBDB.NameToOID(name)
	if oid == ".0.0" {
		if name == "iso" || name == "org" ||
			name == "dod" || name == "internet" ||
			name == ".1" || name == ".1.3" || name == ".1.3.6" {
			return ".1.3.6.1"
		}
		if matched, _ := regexp.MatchString(`\.[0-9.]+`, name); matched {
			return name
		}
	}
	return oid
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
	et := time.Now().Unix() + (3 * 60)
	defer agent.Conn.Close()
	err = agent.Walk(nameToOID(p.Name), func(variable gosnmp.SnmpPDU) error {
		if et < time.Now().Unix() {
			return fmt.Errorf("timeout")
		}
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
						a = []uint8(datastore.PrintMIBStringVal(variable.Value))
					}
					mac := []string{}
					for _, m := range a {
						mac = append(mac, fmt.Sprintf("%02X", m&0x00ff))
					}
					value = strings.Join(mac, ":")
				case "BITS":
					a, ok := variable.Value.([]uint8)
					if !ok {
						a = []uint8(datastore.PrintMIBStringVal(variable.Value))
					}
					hex := []string{}
					ap := []string{}
					bit := 0
					for _, m := range a {
						hex = append(hex, fmt.Sprintf("%02X", m&0x00ff))
						if !p.Raw && mi.Enum != "" {
							for i := 0; i < 8; i++ {
								if (m & 0x80) == 0x80 {
									if n, ok := mi.EnumMap[bit]; ok {
										ap = append(ap, fmt.Sprintf("%s(%d)", n, bit))
									}
								}
								m <<= 1
								bit++
							}
						}
					}
					value = strings.Join(hex, " ")
					if len(ap) > 0 {
						value += " " + strings.Join(ap, " ")
					}
				case "DisplayString":
					value = datastore.PrintMIBStringVal(variable.Value)
					if datastore.MapConf.AutoCharCode {
						value = logger.CheckCharCode(value)
					}
				case "DateAndTime":
					value = datastore.PrintDateAndTime(variable.Value)
				default:
					value = datastore.PrintMIBStringVal(variable.Value)
				}
			} else {
				value = datastore.PrintMIBStringVal(variable.Value)
			}
		case gosnmp.ObjectIdentifier:
			value = datastore.MIBDB.OIDToName(datastore.PrintMIBStringVal(variable.Value))
		case gosnmp.TimeTicks:
			t := gosnmp.ToBigInt(variable.Value).Uint64()
			if p.Raw {
				value = fmt.Sprintf("%d", t)
			} else {
				if t > (24 * 3600 * 100) {
					d := t / (24 * 3600 * 100)
					t -= d * (24 * 3600 * 100)
					value = fmt.Sprintf("%d(%d days, %v)", t, d, time.Duration(t*10*uint64(time.Millisecond)))
				} else {
					value = fmt.Sprintf("%d(%v)", t, time.Duration(t*10*uint64(time.Millisecond)))
				}
			}
		case gosnmp.IPAddress:
			value = datastore.PrintIPAddress(variable.Value)
		default:
			if variable.Type == gosnmp.Integer {
				value = fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Int64())
			} else {
				value = fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Uint64())
			}
			if !p.Raw {
				mi := datastore.FindMIBInfo(name)
				if mi != nil {
					v := int(gosnmp.ToBigInt(variable.Value).Uint64())
					if mi.Enum != "" {
						if vn, ok := mi.EnumMap[v]; ok {
							value += "(" + vn + ")"
						}
					} else {
						if mi.Hint != "" {
							value = datastore.PrintHintedMIBIntVal(int32(v), mi.Hint, variable.Type != gosnmp.Integer)
						}
						if mi.Units != "" {
							value += " " + mi.Units
						}
					}
				}
			}
		}
		ret = append(ret, &mibEnt{
			Name:  name,
			Value: value,
		})
		return nil
	})
	return ret, err
}
