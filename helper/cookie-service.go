package helper

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func DeleteCookie(c echo.Context, cookieName string) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
	c.SetCookie(cookie)
}

func CreateCookie(c echo.Context, cookieName string, value string, maxage int) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxage,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
	c.SetCookie(cookie)
}
