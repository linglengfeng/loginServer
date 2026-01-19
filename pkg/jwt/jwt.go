package jwt

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const jwtsecretKey = "f4X2gkF1sQazWY5A"
const expiredSec = 86400

func EncodeJwt(tokenInfo map[string]any) (string, error) {
	return GenerateJWT(tokenInfo, jwtsecretKey, expiredSec)
}

func DecodeJwt(token string) (jwt.MapClaims, error) {
	return ParseJWT(token, jwtsecretKey)
}

func GenerateJWT(info map[string]any, secretKey string, durationSec int) (string, error) {
	claims := jwt.MapClaims{
		"info": info,
		"exp":  time.Now().Add(time.Duration(durationSec) * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ParseJWT(tokenString, secretKey string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("无效的签名方法：%v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("无效的 JWT")
}
