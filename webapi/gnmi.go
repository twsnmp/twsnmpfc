package webapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/gnmic/pkg/api"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type gnmiWebAPI struct {
	Node *datastore.NodeEnt
}

func getGNMI(c echo.Context) error {
	id := c.Param("id")
	r := gnmiWebAPI{}
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
	if r.Node == nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, r)
}

type gnmiGetReqWebAPI struct {
	NodeID string
	Target string
	Path   string
	Mode   string
}

type gnmiGetEnt struct {
	Path  string
	Value string
	Index string
}

type gnmiCapEnt struct {
	Version   string
	Encodings string
	Models    []*gnmi.ModelData
}

func postGNMI(c echo.Context) error {
	p := new(gnmiGetReqWebAPI)
	if err := c.Bind(p); err != nil {
		log.Printf("gnmi get err=%v", err)
		return echo.ErrBadRequest
	}
	if p.Mode == "Capabilities" {
		r, err := gnmiCap(p)
		if err != nil {
			log.Printf("gnmi cap err=%v", err)
			return echo.ErrBadRequest
		}
		return c.JSON(http.StatusOK, r)
	}
	r, err := gnmiGet(p)
	if err != nil {
		r = append(r, &gnmiGetEnt{
			Path:  "Error",
			Value: err.Error(),
		})
		log.Printf("gnmi get err=%v", err)
		return c.JSON(http.StatusOK, r)
	}
	if len(r) < 1 {
		log.Println("gnmi get not found")
		return echo.ErrNotFound
	}
	return c.JSON(http.StatusOK, r)
}

func gnmiGet(p *gnmiGetReqWebAPI) ([]*gnmiGetEnt, error) {
	ret := []*gnmiGetEnt{}
	n := datastore.GetNode(p.NodeID)
	if n == nil {
		if !strings.HasPrefix(p.NodeID, "NET:") {
			return ret, fmt.Errorf("node not found")
		}
		nt := datastore.GetNetwork(p.NodeID)
		if nt == nil {
			return ret, fmt.Errorf("network not found")
		}
		n = &datastore.NodeEnt{
			IP:       nt.IP,
			User:     nt.User,
			Password: nt.Password,
		}
	}
	tg, err := api.NewTarget(
		api.Name(n.Name),
		api.Address(p.Target),
		api.Username(n.User),
		api.Password(n.Password),
		api.SkipVerify(true),
	)
	if err != nil {
		return ret, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = tg.CreateGNMIClient(ctx)
	if err != nil {
		return ret, err
	}
	defer tg.Close()
	getReq, err := api.NewGetRequest(
		api.Path(p.Path),
		api.Encoding("json_ietf"))
	if err != nil {
		return ret, err
	}
	getResp, err := tg.Get(ctx, getReq)
	if err != nil {
		return ret, err
	}
	for _, not := range getResp.GetNotification() {
		for _, u := range not.GetUpdate() {
			pa := []string{}
			for _, p := range u.Path.Elem {
				pa = append(pa, p.GetName())
			}
			j := u.Val.GetJsonIetfVal()
			var d interface{}
			if err := json.Unmarshal(j, &d); err != nil {
				log.Println(err)
				continue
			}
			path := ""
			if len(pa) > 0 {
				path = "/" + strings.Join(pa, "/")
			}
			ret = append(ret, getPathValue(d, path, "", false)...)
		}
	}
	return ret, nil
}

func getPathValue(d interface{}, path, index string, inArray bool) []*gnmiGetEnt {
	r := []*gnmiGetEnt{}
	switch v := d.(type) {
	case string:
		r = append(r, &gnmiGetEnt{
			Path:  path,
			Value: v,
			Index: index,
		})
		return r
	case float64:
		r = append(r, &gnmiGetEnt{
			Path:  path,
			Value: fmt.Sprintf("%v", v),
			Index: index,
		})
	case bool:
		r = append(r, &gnmiGetEnt{
			Path:  path,
			Value: fmt.Sprintf("%v", v),
			Index: index,
		})
	case map[string]interface{}:
		n := ""
		if in, ok := v["name"]; ok {
			if sn, ok := in.(string); ok {
				n = sn
			}
		}
		for k, vv := range v {
			if inArray && n != "" {
				r = append(r, getPathValue(vv, fmt.Sprintf("%s[name=%s]/%s", path, n, k), index, false)...)
			} else {
				r = append(r, getPathValue(vv, path+"/"+k, index, false)...)
			}
		}
	case []interface{}:
		for i, vv := range v {
			r = append(r, getPathValue(vv, path, fmt.Sprintf("%d", i), true)...)
		}
	default:
		log.Printf("%s=%+v type=%v", path, v, reflect.TypeOf(d))
	}
	return r
}

func gnmiCap(p *gnmiGetReqWebAPI) (*gnmiCapEnt, error) {
	ret := &gnmiCapEnt{}
	n := datastore.GetNode(p.NodeID)
	if n == nil {
		if !strings.HasPrefix(p.NodeID, "NET:") {
			return ret, fmt.Errorf("node not found")
		}
		nt := datastore.GetNetwork(p.NodeID)
		if nt == nil {
			return ret, fmt.Errorf("network not found")
		}
		n = &datastore.NodeEnt{
			IP:       nt.IP,
			User:     nt.User,
			Password: nt.Password,
		}
	}
	tg, err := api.NewTarget(
		api.Name(n.Name),
		api.Address(p.Target),
		api.Username(n.User),
		api.Password(n.Password),
		api.SkipVerify(true),
	)
	if err != nil {
		return ret, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = tg.CreateGNMIClient(ctx)
	if err != nil {
		return ret, err
	}
	defer tg.Close()

	capResp, err := tg.Capabilities(ctx)
	if err != nil {
		return ret, err
	}
	ret.Models = capResp.GetSupportedModels()
	ret.Version = capResp.GetGNMIVersion()
	es := []string{}
	for _, e := range capResp.GetSupportedEncodings() {
		es = append(es, e.String())
	}
	ret.Encodings = strings.Join(es, ",")
	return ret, nil
}
