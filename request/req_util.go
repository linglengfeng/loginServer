package request

import (
	"loginServer/config"
	"loginServer/src/log"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// 响应字段名
	STATUS  = "status"  // 状态码字段（符合 HTTP 标准命名）
	MESSAGE = "message" // 消息字段
	DATA    = "data"    // 数据字段

	// 业务状态码（使用正数，符合 HTTP 状态码范围规范）
	CodeSuccess    = 0    // 成功
	CodeError      = 1001 // 一般错误
	CodeBadRequest = 1002 // 参数错误

	// 状态码对应的默认消息
	MsgSuccess    = "success"
	MsgError      = "failed"
	MsgBadRequest = "parameter error"
)

var (
	// 成功响应模板
	ResponseSuccess = gin.H{
		STATUS:  CodeSuccess,
		MESSAGE: MsgSuccess,
		DATA:    nil,
	}

	// 失败响应模板
	ResponseError = gin.H{
		STATUS:  CodeError,
		MESSAGE: MsgError,
		DATA:    nil,
	}

	// 参数错误响应模板
	ResponseBadRequest = gin.H{
		STATUS:  CodeBadRequest,
		MESSAGE: MsgBadRequest,
		DATA:    nil,
	}
)

// retResponse 统一返回响应（支持消息和数据）
// resp: 响应模板
// message: 消息内容，为空则不设置
// data: 数据内容，为 nil 则不设置
func retResponse(resp gin.H, message string, data any) gin.H {
	if message != "" {
		resp[MESSAGE] = message
	}
	if data != nil {
		resp[DATA] = data
	}
	return resp
}

// JsonBody 解析请求体中的 JSON 数据
func JsonBody(c *gin.Context) (map[string]any, error) {
	var params map[string]any
	if err := c.ShouldBindJSON(&params); err != nil {
		return nil, err
	}
	return params, nil
}

// FormBody 解析请求中的表单数据
func FormBody(c *gin.Context) (map[string]any, error) {
	if err := c.Request.ParseForm(); err != nil {
		return nil, err
	}
	params := make(map[string]any, len(c.Request.PostForm))
	for key, values := range c.Request.PostForm {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params, nil
}

// setupMiddleware 设置中间件
func setupMiddleware(req *gin.Engine) {
	req.Use(func(c *gin.Context) {
		// 处理 OPTIONS 预检请求
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")
			c.Status(http.StatusOK)
			c.Abort()
			return
		}

		// 检查路由访问权限
		allowed, errMsg := shouldDisableRoute(c)
		if !allowed {
			log.Info("request can't used, err:%v", errMsg)
			c.JSON(http.StatusOK, retResponse(ResponseBadRequest, errMsg, nil))
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	})
}

// getClientIP 获取客户端真实IP地址
// 优先级：X-Forwarded-For > X-Real-IP > RemoteAddr
func getClientIP(c *gin.Context) string {
	// 优先从 X-Forwarded-For 获取（经过代理时）
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		if ips := strings.Split(ip, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 从 X-Real-IP 获取
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}

	// 直接从 RemoteAddr 获取
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

// checkIPWhitelist 检查IP是否在白名单中
// 支持精确匹配和CIDR格式（如 192.168.1.0/24）
func checkIPWhitelist(clientIP string, allowedIPs []string) bool {
	clientIPAddr := net.ParseIP(clientIP)
	if clientIPAddr == nil {
		return false
	}

	for _, allowedIP := range allowedIPs {
		allowedIP = strings.TrimSpace(allowedIP)
		if allowedIP == "" {
			continue
		}

		// 精确匹配
		if allowedIP == clientIP {
			return true
		}

		// CIDR格式匹配
		if strings.Contains(allowedIP, "/") {
			_, ipNet, err := net.ParseCIDR(allowedIP)
			if err == nil && ipNet.Contains(clientIPAddr) {
				return true
			}
		}
	}
	return false
}

// shouldDisableRoute 检查路由是否应该被禁用
func shouldDisableRoute(c *gin.Context) (bool, string) {
	fullpath := c.FullPath()
	allowed, errMsg := checkLimitApi(fullpath, c)
	if !allowed {
		return false, errMsg
	}
	return true, ""
}

// checkLimitApi 检查 API 访问限制
func checkLimitApi(fullpath string, c *gin.Context) (bool, string) {
	// 开发环境不限制
	if config.Config.GetString("server_mod") == "dev" {
		return true, ""
	}

	// 从全局路由表获取路由信息
	route, exists := routeCache[fullpath]
	if !exists {
		return true, ""
	}

	// 检查调试接口限制
	if route.IsDebug {
		return false, "生产环境不允许访问调试接口"
	}

	// 检查IP白名单
	if route.ApiGroup != "" {
		allowedIPs := getAllowedIPsByGroup(route.ApiGroup)
		clientIP := getClientIP(c)
		if !checkIPWhitelist(clientIP, allowedIPs) {
			log.Info("IP不允许访问: %s, 路径: %s", clientIP, fullpath)
			return false, "IP不在白名单中"
		}
	}

	return true, ""
}

// getAllowedIPsByGroup 根据API分组获取IP白名单
// 如果配置不存在或为空，返回 nil 表示不限制IP
func getAllowedIPsByGroup(apiGroup string) []string {
	configKey := "ip_whitelist." + apiGroup
	ips := config.Config.GetStringSlice(configKey)
	if len(ips) == 0 {
		return nil
	}
	return ips
}
