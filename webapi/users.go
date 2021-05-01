package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
)

func getUsers(c echo.Context) error {
	r := []*datastore.UserEnt{}
	datastore.ForEachUsers(func(u *datastore.UserEnt) bool {
		if u.ValidScore {
			r = append(r, u)
		}
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteUser(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("users")
	} else {
		datastore.DeleteUser(id)
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func resetUsers(c echo.Context) error {
	report.ResetUsersScore()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
