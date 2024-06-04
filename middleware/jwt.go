package middleware

import (
	"HospitalManager/security"
	"github.com/golang-jwt/jwt"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

func SetJWTHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookies := c.Cookies()
		tokenExist := false
		for _, cookie := range cookies {
			if cookie.Name == "jwt" || cookie.Name == "refresh-token" {
				tokenExist = true
				break
			}
		}
		if tokenExist {
			token, _ := c.Cookie("jwt")
			c.Request().Header.Set("Authorization", "Bearer "+token.Value)
			return next(c)
		}
		return c.JSON(http.StatusNotFound, "Token not found 2")
	}
}

func JWTMiddleware() echo.MiddlewareFunc {
	config := echojwt.Config{
		SigningKey: []byte(security.SECRET_KEY),
	}
	return echojwt.WithConfig(config)
}

func ValidateAndExtractClaims(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		if tokenString == "" {
			return c.JSON(http.StatusNotFound, "Token not found 2")
		}
		token, err := security.ValidateToken(tokenString)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, "Token invalid or expired")
		}
		if token.Valid {
			claims := token.Claims.(jwt.MapClaims)
			c.Set("Role", claims["role"])
			c.Set("Userid", claims["userid"])
			return next(c)
		}
		return c.JSON(http.StatusUnauthorized, "Token invalid or expired")
	}
}
