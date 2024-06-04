package security

import (
	"HospitalManager/model"
	"errors"
	"github.com/golang-jwt/jwt"
	"time"
)

const SECRET_KEY = "sonnvt2002@gmail.com"

func GenToken(userid string, role string, expiryTime time.Duration) (string, error) {
	claims := &model.CustomClaims{
		Userid: userid,
		Role:   role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiryTime).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	result, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}

	return result, nil
}

func ValidateToken(token string) (*jwt.Token, error) {
	claims, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Error")
		}
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func GenAccessTokenFromRefreshToken(refreshToken string) (string, error) {
	rfToken, err := ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}
	if !rfToken.Valid {
		return "", errors.New("Token unvalid")
	}
	claims := rfToken.Claims.(jwt.MapClaims)
	userid := claims["userid"].(string)
	role := claims["role"].(string)
	newAccessToken, err := GenToken(userid, role, time.Hour)
	if err != nil {
		return "", err
	}
	return newAccessToken, nil
}
