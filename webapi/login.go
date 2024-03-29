// Package webapi : WEB API
package webapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/security"
)

type loginWebAPI struct {
	UserID   string `json:"UserID" form:"UserID" query:"UserID"`
	Password string `json:"Password" form:"Password" query:"Password"`
}

func login(c echo.Context) error {
	le := new(loginWebAPI)
	if err := c.Bind(le); err != nil {
		return echo.ErrUnauthorized
	}
	api := c.Get("api").(*WebAPI)
	// パスワード認証
	if le.UserID != datastore.MapConf.UserID ||
		!security.PasswordVerify(datastore.MapConf.Password, le.Password) {
		return echo.ErrUnauthorized
	}

	// トークン作成
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["userid"] = le.UserID
	if api.Timeout > 0 {
		claims["exp"] = time.Now().Add(time.Hour * time.Duration(api.Timeout)).Unix()
	} else {
		claims["exp"] = time.Now().Add(time.Hour * 24 * 365 * 10).Unix()
	}

	t, err := token.SignedString([]byte(api.Password))
	if err != nil {
		return err
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("%sからログインしました", c.RealIP()),
	})
	return c.JSON(http.StatusOK, map[string]string{
		"token": t,
	})
}

type meWebAPI struct {
	ID     int    `json:"id"`
	UserID string `json:"userid"`
}

func getMe(c echo.Context) error {
	r := new(meWebAPI)
	r.ID = 1
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	r.UserID = claims["userid"].(string)
	return c.JSON(http.StatusOK, r)
}

func getVersion(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	return c.JSON(http.StatusOK, map[string]string{"Version": api.Version})
}
