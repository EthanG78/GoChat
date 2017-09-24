package middleware

import (
	"github.com/labstack/echo"
	"strings"
	"net/http"
	"log"
)

func ServerInfo(pipe echo.HandlerFunc) echo.HandlerFunc{
	return func (c echo.Context) error{
		c.Response().Header().Set(echo.HeaderServer, "GoChat/1.0")
		c.Response().Header().Set("Author", "Ethan Garnier")
		return pipe(c)
	}
}

func checkCookies (pipe echo.HandlerFunc) echo.HandlerFunc{
	return func (c echo.Context) error{
		cookie, err := c.Cookie("session_id")
		if err != nil{
			if strings.Contains(err.Error(), "named cookie not present"){
				return c.String(http.StatusUnauthorized, "Seems like the cookie monster was here..\n You do not have any cookies")
			}

			log.Printf("Error: %v \n", err)
			return err


		}

		if cookie.Value == CookieVal{
			return pipe(c)
		}

		return c.String(http.StatusUnauthorized, "You do not have the correct cookie")
	}
}