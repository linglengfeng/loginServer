package request

import (
	"fmt"
	"loginServer/pkg/crypto"
	"loginServer/pkg/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// handle_encrypt 加密处理函数，对传入的信息进行加密
// POST /loginServer/test/encrypt
func handle_encrypt(c *gin.Context) {
	info := c.PostForm("info")
	if info == "" {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "", nil))
		return
	}

	infostr, err := crypto.Encrypt(info)
	if err != nil {
		c.JSON(http.StatusOK, retResponse(CodeError, fmt.Sprintf("加密失败: %s", err.Error()), nil))
		return
	}
	c.JSON(http.StatusOK, retResponse(CodeSuccess, "", infostr))
}

// handle_decrypt 解密处理函数，对传入的加密信息进行解密
// POST /loginServer/test/decrypt
func handle_decrypt(c *gin.Context) {
	info := c.PostForm("info")
	if info == "" {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "", nil))
		return
	}
	infostr, err := crypto.Decrypt(info)
	if err != nil {
		c.JSON(http.StatusOK, retResponse(CodeError, fmt.Sprintf("解密失败: %s", err.Error()), nil))
		return
	}
	c.JSON(http.StatusOK, retResponse(CodeSuccess, "", infostr))
}

// handle_encodejwt JWT 编码处理函数，将信息编码为 JWT Token
// POST /loginServer/test/encodeJwt
func handle_encodejwt(c *gin.Context) {
	info := c.PostForm("info")
	if info == "" {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "", nil))
		return
	}
	mapinfo := map[string]any{"token": info}
	token, err := jwt.EncodeJwt(mapinfo)
	if err != nil {
		c.JSON(http.StatusOK, retResponse(CodeError, fmt.Sprintf("JWT编码失败: %s", err.Error()), nil))
		return
	}
	c.JSON(http.StatusOK, retResponse(CodeSuccess, "", token))
}

// handle_decodejwt JWT 解码处理函数，将 JWT Token 解码为原始信息
// POST /loginServer/test/decodeJwt
func handle_decodejwt(c *gin.Context) {
	info := c.PostForm("info")
	if info == "" {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "", nil))
		return
	}
	tokeninfo, err := jwt.DecodeJwt(info)
	if err != nil {
		c.JSON(http.StatusOK, retResponse(CodeError, fmt.Sprintf("JWT解码失败: %s", err.Error()), nil))
		return
	}
	c.JSON(http.StatusOK, retResponse(CodeSuccess, "", tokeninfo))
}
