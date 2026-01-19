package request

import (
	"fmt"
	"loginServer/pkg/crypto"
	"loginServer/pkg/jwt"
	"loginServer/src/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 获取服务器列表处理函数
func handle_getServerList(c *gin.Context) {
	servers, err := GetServerList()
	if err != nil {
		log.Error("handle_getServerList failed, err: %v", err)
		c.JSON(http.StatusOK, retResponse(ResponseError, "获取服务器列表失败", nil))
		return
	}
	c.JSON(http.StatusOK, retResponse(ResponseSuccess, "", servers))
}

// 游戏服上报服务器列表处理函数
func handle_reportServerList(c *gin.Context) {
	servers, err := GetServerList()
	if err != nil {
		log.Error("handle_reportServerList failed, err: %v", err)
		c.JSON(http.StatusOK, retResponse(ResponseError, "获取服务器列表失败", nil))
		return
	}
	c.JSON(http.StatusOK, retResponse(ResponseSuccess, "", servers))
}

// 游戏服服务器状态变更处理函数
func handle_changeServerState(c *gin.Context) {
	servers, err := GetServerList()
	if err != nil {
		log.Error("handle_changeServerState failed, err: %v", err)
		c.JSON(http.StatusOK, retResponse(ResponseError, "获取服务器列表失败", nil))
		return
	}
	c.JSON(http.StatusOK, retResponse(ResponseSuccess, "", servers))
}

// 测试接口处理函数，能够测试所有的返回情况
// 参数说明：
//   - case: 测试场景（可选值：success, error, param_error, all）
//   - with_message: 是否包含消息（true/false，默认 false）
//   - with_data: 是否包含数据（true/false，默认 true）
//
// 示例：
//   - case=success&with_data=true -> 成功响应 + 数据
//   - case=success&with_message=true&with_data=true -> 成功响应 + 消息 + 数据
//   - case=error&with_message=true -> 失败响应 + 消息
//   - case=param_error&with_message=true -> 参数错误响应 + 消息
//   - case=all -> 返回所有测试场景的说明
func handle_test(c *gin.Context) {
	params, err := FormBody(c)
	if err != nil {
		c.JSON(http.StatusOK, retResponse(ResponseBadRequest, fmt.Sprintf("解析表单参数失败: %v", err), nil))
		return
	}

	// 获取测试参数
	testCase := c.PostForm("case")
	withMessage := c.PostForm("with_message") == "true"
	withData := c.PostForm("with_data") != "false" // 默认为 true

	// 测试数据
	testData := map[string]any{
		"test_key":    "test_value",
		"test_number": 123,
		"test_array":  []string{"a", "b", "c"},
		"received":    params,
	}
	testMessage := "这是测试消息"

	// 根据 case 参数返回不同的响应
	switch testCase {
	case "success":
		// 成功响应
		message := ""
		data := any(nil)
		if withMessage {
			message = testMessage
		}
		if withData {
			data = testData
		}
		c.JSON(http.StatusOK, retResponse(ResponseSuccess, message, data))
		return

	case "error":
		// 失败响应
		message := ""
		data := any(nil)
		if withMessage {
			message = "测试失败消息"
		}
		if withData {
			data = testData
		}
		c.JSON(http.StatusOK, retResponse(ResponseError, message, data))
		return

	case "param_error":
		// 参数错误响应
		message := ""
		data := any(nil)
		if withMessage {
			message = "测试参数错误消息"
		}
		if withData {
			data = testData
		}
		c.JSON(http.StatusOK, retResponse(ResponseBadRequest, message, data))
		return

	case "all":
		// 返回所有测试场景说明
		allCases := map[string]any{
			"说明": "测试接口 - 可以测试所有返回情况",
			"参数": map[string]string{
				"case":         "测试场景：success(成功)、error(失败)、param_error(参数错误)、all(查看说明)",
				"with_message": "是否包含消息：true/false（默认 false）",
				"with_data":    "是否包含数据：true/false（默认 true）",
			},
			"测试场景": []map[string]string{
				{"场景": "成功响应 + 数据", "参数": "case=success&with_data=true"},
				{"场景": "成功响应 + 消息 + 数据", "参数": "case=success&with_message=true&with_data=true"},
				{"场景": "成功响应 + 消息", "参数": "case=success&with_message=true&with_data=false"},
				{"场景": "失败响应 + 消息", "参数": "case=error&with_message=true"},
				{"场景": "失败响应 + 消息 + 数据", "参数": "case=error&with_message=true&with_data=true"},
				{"场景": "参数错误响应 + 消息", "参数": "case=param_error&with_message=true"},
				{"场景": "参数错误响应 + 消息 + 数据", "参数": "case=param_error&with_message=true&with_data=true"},
			},
		}
		c.JSON(http.StatusOK, retResponse(ResponseSuccess, "", allCases))
		return

	default:
		// 默认返回成功响应 + 接收到的参数
		c.JSON(http.StatusOK, retResponse(ResponseSuccess, "", params))
	}
}

// 加密处理函数，对传入的信息进行加密
func handle_encrypt(c *gin.Context) {
	info := c.PostForm("info")
	if info == "" {
		c.JSON(http.StatusOK, ResponseBadRequest)
		return
	}

	infostr, err := crypto.Encrypt(info)
	if err != nil {
		c.JSON(http.StatusOK, retResponse(ResponseError, fmt.Sprintf("加密失败: %s", err.Error()), nil))
		return
	}
	c.JSON(http.StatusOK, retResponse(ResponseSuccess, "", infostr))
}

// 解密处理函数，对传入的加密信息进行解密
func handle_decrypt(c *gin.Context) {
	info := c.PostForm("info")
	if info == "" {
		c.JSON(http.StatusOK, ResponseBadRequest)
		return
	}
	infostr, err := crypto.Decrypt(info)
	if err != nil {
		c.JSON(http.StatusOK, retResponse(ResponseError, fmt.Sprintf("解密失败: %s", err.Error()), nil))
		return
	}
	c.JSON(http.StatusOK, retResponse(ResponseSuccess, "", infostr))
}

// JWT 编码处理函数，将信息编码为 JWT Token
func handle_encodejwt(c *gin.Context) {
	info := c.PostForm("info")
	if info == "" {
		c.JSON(http.StatusOK, ResponseBadRequest)
		return
	}
	mapinfo := map[string]any{"token": info}
	token, err := jwt.EncodeJwt(mapinfo)
	if err != nil {
		c.JSON(http.StatusOK, retResponse(ResponseError, fmt.Sprintf("JWT编码失败: %s", err.Error()), nil))
		return
	}
	c.JSON(http.StatusOK, retResponse(ResponseSuccess, "", token))
}

// JWT 解码处理函数，将 JWT Token 解码为原始信息
func handle_decodejwt(c *gin.Context) {
	info := c.PostForm("info")
	if info == "" {
		c.JSON(http.StatusOK, ResponseBadRequest)
		return
	}
	tokeninfo, err := jwt.DecodeJwt(info)
	if err != nil {
		c.JSON(http.StatusOK, retResponse(ResponseError, fmt.Sprintf("JWT解码失败: %s", err.Error()), nil))
		return
	}
	c.JSON(http.StatusOK, retResponse(ResponseSuccess, "", tokeninfo))
}
