package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

func ServerInfo(pipe echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderServer, "GoChat/1.0")
		c.Response().Header().Set("Author", "Ethan Garnier")
		return pipe(c)
	}
}

func CheckCookies(pipe echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("session_id")
		if err != nil {
			if strings.Contains(err.Error(), "named cookie not present") {
				return c.String(http.StatusUnauthorized, "Seems like the cookie monster was here..\n You do not have any cookies")
			}

			log.Printf("Error: %v \n", err)
			return err

		}

		if cookie.Name == "session_id" {
			return pipe(c)
		}

		return c.String(http.StatusUnauthorized, "You don't have any cookies")
	}
}
