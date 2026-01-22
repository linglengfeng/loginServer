package request

import (
	"github.com/gin-gonic/gin"
)

// RouteMethod HTTP 方法类型
type RouteMethod string

const (
	MethodGET     RouteMethod = "GET"
	MethodPOST    RouteMethod = "POST"
	MethodPUT     RouteMethod = "PUT"
	MethodDELETE  RouteMethod = "DELETE"
	MethodPATCH   RouteMethod = "PATCH"
	MethodOPTIONS RouteMethod = "OPTIONS"
	MethodHEAD    RouteMethod = "HEAD"
)

const (
	ApiGroupOut   = "out"
	ApiGroupSgame = "sgame"
	ApiGroupTest  = "test"
)

// Route 路由信息结构体
type Route struct {
	Path     string
	Method   RouteMethod
	Handler  gin.HandlerFunc
	IsDebug  bool   // 是否为调试接口（仅在开发环境可用）
	ApiGroup string // API分组：out(对外接口)、sgame(游戏服接口)、test(测试接口)
}

// 路由路径常量
const (
	// out 分组
	PathGetServerList       = "/loginServer/getServerList"       // 获取服务器列表
	PathGetPlayerServerList = "/loginServer/getPlayerServerList" // 获取玩家服务器列表

	// sgame 分组
	PathTest              = "/loginServer/test"              // 测试接口（GET）
	PathTestPost          = "/loginServer/testPost"          // 测试接口（POST）
	PathReportServerList  = "/loginServer/reportServerList"  // 游戏服上报服务器列表
	PathChangeServerState = "/loginServer/changeServerState" // 游戏服服务器状态变更
	PathSetUserHistory    = "/loginServer/setUserHistory"    // 玩家信息设置

	// test 分组
	PathEncrypt   = "/encrypt"   // 加密接口
	PathDecrypt   = "/decrypt"   // 解密接口
	PathEncodeJwt = "/encodeJwt" // 编码JWT
	PathDecodeJwt = "/decodeJwt" // 解码JWT
)

// routeCache 路由映射表（path -> Route），在包初始化时构建一次，可供整个包复用。
var routeCache = map[string]Route{
	// out 分组
	PathGetServerList:       {Path: PathGetServerList, Method: MethodGET, Handler: handle_getServerList, IsDebug: false, ApiGroup: ApiGroupOut},
	PathGetPlayerServerList: {Path: PathGetPlayerServerList, Method: MethodGET, Handler: handle_getPlayerServerList, IsDebug: false, ApiGroup: ApiGroupOut},

	// sgame 分组
	PathTest:              {Path: PathTest, Method: MethodGET, Handler: handle_test, IsDebug: true, ApiGroup: ApiGroupSgame},
	PathTestPost:          {Path: PathTestPost, Method: MethodPOST, Handler: handle_testPost, IsDebug: true, ApiGroup: ApiGroupSgame},
	PathReportServerList:  {Path: PathReportServerList, Method: MethodPOST, Handler: handle_reportServerList, IsDebug: false, ApiGroup: ApiGroupSgame},
	PathChangeServerState: {Path: PathChangeServerState, Method: MethodPOST, Handler: handle_changeServerState, IsDebug: false, ApiGroup: ApiGroupSgame},
	PathSetUserHistory:    {Path: PathSetUserHistory, Method: MethodPOST, Handler: handle_SetUserHistory, IsDebug: false, ApiGroup: ApiGroupSgame},

	// test 分组
	PathEncrypt:   {Path: PathEncrypt, Method: MethodPOST, Handler: handle_encrypt, IsDebug: true, ApiGroup: ApiGroupTest},
	PathDecrypt:   {Path: PathDecrypt, Method: MethodPOST, Handler: handle_decrypt, IsDebug: true, ApiGroup: ApiGroupTest},
	PathEncodeJwt: {Path: PathEncodeJwt, Method: MethodPOST, Handler: handle_encodejwt, IsDebug: true, ApiGroup: ApiGroupTest},
	PathDecodeJwt: {Path: PathDecodeJwt, Method: MethodPOST, Handler: handle_decodejwt, IsDebug: true, ApiGroup: ApiGroupTest},
}

// methodHandlers HTTP 方法到注册函数的映射
var methodHandlers = map[RouteMethod]func(*gin.Engine, string, ...gin.HandlerFunc) gin.IRoutes{
	MethodGET:     (*gin.Engine).GET,
	MethodPOST:    (*gin.Engine).POST,
	MethodPUT:     (*gin.Engine).PUT,
	MethodDELETE:  (*gin.Engine).DELETE,
	MethodPATCH:   (*gin.Engine).PATCH,
	MethodOPTIONS: (*gin.Engine).OPTIONS,
	MethodHEAD:    (*gin.Engine).HEAD,
}

// request 注册所有路由到 Gin 引擎
func request(req *gin.Engine) {
	setupMiddleware(req)
	for _, route := range routeCache {
		if handler, exists := methodHandlers[route.Method]; exists {
			handler(req, route.Path, route.Handler)
		}
	}
}
