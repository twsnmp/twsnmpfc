package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func getWinEventID(c echo.Context) error {
	r := []*datastore.WinEventIDEnt{}
	datastore.ForEachWinEventID(func(e *datastore.WinEventIDEnt) bool {
		r = append(r, e)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteWinEventID(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("winEventID")
	} else {
		datastore.DeleteReport("winEventID", id)
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getWinLogon(c echo.Context) error {
	r := []*datastore.WinLogonEnt{}
	datastore.ForEachWinLogon(func(e *datastore.WinLogonEnt) bool {
		r = append(r, e)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteWinLogon(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("winLogon")
	} else {
		datastore.DeleteReport("winLogon", id)
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getWinAccount(c echo.Context) error {
	r := []*datastore.WinAccountEnt{}
	datastore.ForEachWinAccount(func(e *datastore.WinAccountEnt) bool {
		r = append(r, e)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteWinAccount(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("winAccount")
	} else {
		datastore.DeleteReport("winAccount", id)
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getWinKerberos(c echo.Context) error {
	r := []*datastore.WinKerberosEnt{}
	datastore.ForEachWinKerberos(func(e *datastore.WinKerberosEnt) bool {
		r = append(r, e)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteWinKerberos(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("winKerberos")
	} else {
		datastore.DeleteReport("winKerberos", id)
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getWinPrivilege(c echo.Context) error {
	r := []*datastore.WinPrivilegeEnt{}
	datastore.ForEachWinPrivilege(func(e *datastore.WinPrivilegeEnt) bool {
		r = append(r, e)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteWinPrivilege(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("winPrivilege")
	} else {
		datastore.DeleteReport("winPrivilege", id)
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getWinProcess(c echo.Context) error {
	r := []*datastore.WinProcessEnt{}
	datastore.ForEachWinProcess(func(e *datastore.WinProcessEnt) bool {
		r = append(r, e)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteWinProcess(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("winProcess")
	} else {
		datastore.DeleteReport("winProcess", id)
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getWinTask(c echo.Context) error {
	r := []*datastore.WinTaskEnt{}
	datastore.ForEachWinTask(func(s *datastore.WinTaskEnt) bool {
		r = append(r, s)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteWinTask(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("winTask")
	} else {
		datastore.DeleteReport("winTask", id)
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
