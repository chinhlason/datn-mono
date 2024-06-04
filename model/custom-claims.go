package model

import "github.com/golang-jwt/jwt"

type CustomClaims struct {
	Userid string `json:"userid"`
	Role   string `json:"role"`
	jwt.StandardClaims
}
