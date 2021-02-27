package webapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type snmpTrapFilter struct {
	StartDate   string
	StartTime   string
	EndDate     string
	EndTime     string
	FromAddress string
	TrapType    string
	Variables   string
}

type snmpTrapWebAPI struct {
	Time        int64
	FromAddress string
	TrapType    string
	Variables   string
}

var trapOidRegexp = regexp.MustCompile(`snmpTrapOID.0=(\S+)`)

func postSnmpTrap(c echo.Context) error {
	r := []*snmpTrapWebAPI{}
	var err error
	filter := new(snmpTrapFilter)
	if err = c.Bind(filter); err != nil {
		log.Printf("postSyslog err=%v", err)
		return echo.ErrBadRequest
	}
	fromAddressFilter := makeStringFilter(filter.FromAddress)
	trapTypeFilter := makeStringFilter(filter.TrapType)
	variablesFilter := makeStringFilter(filter.Variables)
	st := makeTimeFilter(filter.StartDate, filter.StartTime, 24)
	et := makeTimeFilter(filter.EndDate, filter.EndTime, 0)
	i := 0
	datastore.ForEachLog(st, et, "trap", func(l *datastore.LogEnt) bool {
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			log.Printf("postSyslog err=%v", err)
			return true
		}
		var ok bool
		re := new(snmpTrapWebAPI)
		if re.FromAddress, ok = sl["FromAddress"].(string); !ok {
			log.Printf("postSyslog no FromAddress")
			return true
		}
		if re.Variables, ok = sl["Variables"].(string); !ok {
			log.Printf("postSyslog no Variables")
			return true
		}
		var ent string
		if ent, ok = sl["Enterprise"].(string); !ok || ent == "" {
			a := trapOidRegexp.FindStringSubmatch(re.Variables)
			if len(a) > 1 {
				re.TrapType = a[1]
			} else {
				re.TrapType = ""
			}
		} else {
			var gen float64
			if gen, ok = sl["GenericTrap"].(float64); !ok {
				log.Printf("postSyslog no GenericTrap")
				return true
			}
			var spe float64
			if spe, ok = sl["SpecificTrap"].(float64); !ok {
				log.Printf("postSyslog no SpecificTrap")
				return true
			}
			re.TrapType = fmt.Sprintf("%s:%d:%d", ent, int(gen), int(spe))
		}
		re.Time = l.Time
		if fromAddressFilter != nil && !fromAddressFilter.Match([]byte(re.FromAddress)) {
			return true
		}
		if variablesFilter != nil && !variablesFilter.Match([]byte(re.Variables)) {
			return true
		}
		if trapTypeFilter != nil && !trapTypeFilter.Match([]byte(re.TrapType)) {
			return true
		}
		r = append(r, re)
		i++
		return i <= datastore.MapConf.LogDispSize
	})
	return c.JSON(http.StatusOK, r)
}
