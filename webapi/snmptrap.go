package webapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

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
		return echo.ErrBadRequest
	}
	fromAddressFilter := makeStringFilter(filter.FromAddress)
	trapTypeFilter := makeStringFilter(filter.TrapType)
	variablesFilter := makeStringFilter(filter.Variables)
	st := makeStartTimeFilter(filter.StartDate, filter.StartTime)
	et := makeEndTimeFilter(filter.EndDate, filter.EndTime)
	datastore.ForEachLogReverse(st, et, "trap", func(l *datastore.LogEnt) bool {
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			return true
		}
		var ok bool
		re := new(snmpTrapWebAPI)
		if fa, ok := sl["FromAddress"].(string); !ok {
			return true
		} else {
			a := strings.SplitN(fa, ":", 2)
			if len(a) == 2 {
				re.FromAddress = a[0]
				n := datastore.FindNodeFromIP(a[0])
				if n != nil {
					re.FromAddress += "(" + n.Name + ")"
				}
			} else {
				re.FromAddress = fa
			}
		}

		if re.Variables, ok = sl["Variables"].(string); !ok {
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
				return true
			}
			var spe float64
			if spe, ok = sl["SpecificTrap"].(float64); !ok {
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
		return len(r) <= datastore.MapConf.LogDispSize
	})
	// 逆順にする
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return c.JSON(http.StatusOK, r)
}
