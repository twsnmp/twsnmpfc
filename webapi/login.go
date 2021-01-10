// Package webapi : WEB API
package webapi

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

type loginEnt struct {
	Name     string `json:"name" form:"name" query:"name"`
	Password string `json:"password" form:"password" query:"password"`
}

func login(c echo.Context) error {
	le := new(loginEnt)
	if err := c.Bind(le); err != nil {
		return echo.ErrUnauthorized
	}
	api := c.Get("api").(*WebAPI)
	// とりあえずのパスワード認証
	if le.Name != "test" || le.Password != "test" {
		return echo.ErrUnauthorized
	}

	// トークン作成
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = le.Name
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	t, err := token.SignedString([]byte(api.Password))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{
		"token": t,
	})
}

type testResEnt struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	NickName string `json:"nickname"`
}

func apiTest(c echo.Context) error {
	r := new(testResEnt)
	r.ID = 1
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	r.Name = claims["name"].(string)
	r.NickName = "Test User"
	return c.JSON(http.StatusOK, r)
}
