package main

import (
	"net/http"

	"github.com/PCCSuite/PCCSamba/SambaAPI/lib/db"
	"github.com/PCCSuite/PCCSamba/SambaAPI/lib/handler"
	"github.com/labstack/echo/v4"
)

func main() {
	db.InitDB()

	e := echo.New()

	e.GET("/", root)
	e.GET("/ping", ping)
	e.GET("/getPassword", handler.GetPassword)
	e.POST("/setPassword", handler.SetPassword)

	e.Start(":8080")
}

func root(c echo.Context) error {
	return c.String(http.StatusNotFound, "This is samba API server")
}

func ping(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}
