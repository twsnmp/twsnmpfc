package webapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func TestGetPublicMap(t *testing.T) {
	e := echo.New()

	// 1. 公開が無効な場合 -> 403
	datastore.MapConf.EnablePublicDashboard = false
	datastore.MapConf.PublicDashboardKey = "testkey123"

	req := httptest.NewRequest(http.MethodGet, "/public/api/map/testkey123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("key")
	c.SetParamValues("testkey123")

	err := getPublicMap(c)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if he, ok := err.(*echo.HTTPError); ok {
		if he.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden, got %d", he.Code)
		}
	} else {
		t.Errorf("expected HTTPError, got %v", err)
	}

	// 2. 公開は有効だがキーが不一致の場合 -> 403
	datastore.MapConf.EnablePublicDashboard = true
	req = httptest.NewRequest(http.MethodGet, "/public/api/map/wrongkey", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("key")
	c.SetParamValues("wrongkey")

	err = getPublicMap(c)
	if err == nil {
		t.Errorf("expected error for wrong key, got nil")
	}

	// 3. 公開が有効でキーが一致する場合 -> 200 & 機密情報がサニタイズされていること
	datastore.MapConf.UserID = "admin_user"
	datastore.MapConf.Password = "secret_pass"
	datastore.MapConf.Community = "secret_community"

	req = httptest.NewRequest(http.MethodGet, "/public/api/map/testkey123", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("key")
	c.SetParamValues("testkey123")

	err = getPublicMap(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rec.Code)
	}

	var resp mapWebAPI
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.MapConf.UserID != "" || resp.MapConf.Password != "" || resp.MapConf.Community != "" {
		t.Errorf("sensitive MapConf information was not sanitized: %+v", resp.MapConf)
	}
	if resp.MapConf.PublicDashboardKey != "" {
		t.Errorf("PublicDashboardKey should not be exposed in public response")
	}
	if len(resp.Logs) != 0 {
		t.Errorf("Logs should be empty in public map response")
	}
	for _, n := range resp.Nodes {
		if n.IP != "" || n.IPv6 != "" || n.MAC != "" || n.Community != "" || n.Password != "" {
			t.Errorf("node sensitive info was not sanitized: %+v", n)
		}
	}
	for _, nw := range resp.Networks {
		if nw.IP != "" || nw.Community != "" || nw.Password != "" {
			t.Errorf("network sensitive info was not sanitized: %+v", nw)
		}
	}
}
