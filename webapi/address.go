package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/report"
)

func getAddressInfo(c echo.Context) error {
	addr := c.Param("addr")
	return c.JSON(http.StatusOK, report.GetAddressInfo(addr))
}
