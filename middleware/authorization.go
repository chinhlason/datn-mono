package middleware

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func IsADMIN(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		Role := c.Get("Role")
		if Role != "HEAD_DOCTOR" {
			return c.JSON(http.StatusForbidden, "You do not have permission to access this resource")
		}
		return next(c)
	}
}
