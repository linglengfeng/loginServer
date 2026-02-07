package request

import (
	"loginServer/config"
	"loginServer/src/log"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 统一响应结构
type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	// 业务状态码（使用正数，符合 HTTP 状态码范围规范）
	CodeSuccess    = 0    // 成功
	CodeError      = 1001 // 一般错误
	CodeBadRequest = 1002 // 参数错误

	// 状态码对应的默认消息
	MsgSuccess    = "success"
	MsgError      = "failed"
	MsgBadRequest = "parameter error"
)

// retResponse 统一返回响应（支持消息和数据）
// code: 业务状态码
// message: 调用方自定义消息，为空时使用该 code 对应的默认消息
// data: 数据内容，为 nil 则不设置
func retResponse(code int, message string, data any) Response {
	resp := Response{
		Status: code,
	}

	// 优先使用自定义 message，没有就用该 code 对应的默认消息
	if message == "" {
		switch code {
		case CodeSuccess:
			resp.Message = MsgSuccess
		case CodeError:
			resp.Message = MsgError
		case CodeBadRequest:
			resp.Message = MsgBadRequest
		default:
			resp.Message = ""
		}
	} else {
		resp.Message = message
	}

	if data != nil {
		resp.Data = data
	}
	return resp
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
			c.AbortWithStatusJSON(http.StatusForbidden, retResponse(CodeBadRequest, errMsg, nil))
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
// 参数 allowedIPs 的含义：
//   - nil: 配置不存在，返回 true（允许所有 IP 访问）
//   - []: 配置存在但为空列表，返回 false（不允许任何 IP 访问）
//   - [ip1, ip2, ...]: 检查 IP 是否在列表中
func checkIPWhitelist(clientIP string, allowedIPs []string) bool {
	// 如果 allowedIPs 为 nil，表示配置不存在，允许所有 IP 访问
	if allowedIPs == nil {
		return true
	}

	// 如果 allowedIPs 为空列表，表示配置存在但为空，不允许任何 IP 访问
	if len(allowedIPs) == 0 {
		return false
	}

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
	// 在中间件中，FullPath() 可能为空（路由还未匹配），使用 Request.URL.Path
	fullpath := c.FullPath()
	if fullpath == "" {
		fullpath = c.Request.URL.Path
	}
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
			return false, "IP不在白名单中"
		}
	}

	return true, ""
}

// getAllowedIPsByGroup 根据API分组获取IP白名单
// 返回值和含义：
//   - nil: 配置不存在，表示不限制IP（允许所有访问）
//   - []: 配置存在但为空列表，表示不允许任何IP访问
//   - [ip1, ip2, ...]: 配置了白名单IP列表
func getAllowedIPsByGroup(apiGroup string) []string {
	// 使用动态白名单管理器
	return GetAllowedIPsByGroup(apiGroup)
}
