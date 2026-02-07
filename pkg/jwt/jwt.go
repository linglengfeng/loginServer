package jwt

import (
	"fmt"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const defaultJWTSecretKey = "f4X2gkF1sQazWY5A"
const expiredSec = 86400

func EncodeJwt(tokenInfo map[string]any) (string, error) {
	return GenerateJWT(tokenInfo, getJWTSecretKey(), expiredSec)
}

func DecodeJwt(token string) (jwt.MapClaims, error) {
	return ParseJWT(token, getJWTSecretKey())
}

func getJWTSecretKey() string {
	// 允许通过环境变量覆盖（便于部署时配置，不要硬编码在代码里）
	if v := os.Getenv("LOGIN_SERVER_JWT_SECRET"); v != "" {
		return v
	}
	return defaultJWTSecretKey
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
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			// 严格限制算法，避免 alg 混淆
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("无效的签名方法：%v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}
	if token == nil || !token.Valid {
		return nil, fmt.Errorf("无效的 JWT")
	}

	// v5 会自动校验 exp（如存在），这里直接返回 claims
	return claims, nil
}
