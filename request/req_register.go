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

// Route 路由信息结构体
type Route struct {
	Path     string
	Method   RouteMethod
	Handler  gin.HandlerFunc
	IsDebug  bool   // 是否为调试接口（仅在开发环境可用）
	ApiGroup string // API分组：out(对外接口)、sgame(游戏服接口)、test(测试接口)
}

const (
	ApiGroupTest        = "test"
	ApiGroupOut         = "out"
	ApiGroupSgame       = "sgame"
	ApiGroupAdminServer = "adminServer"
)

// 路由路径常量
const (
	// test 分组
	PathEncrypt   = "/loginServer/test/encrypt"   // 加密接口
	PathDecrypt   = "/loginServer/test/decrypt"   // 解密接口
	PathEncodeJwt = "/loginServer/test/encodeJwt" // 编码JWT
	PathDecodeJwt = "/loginServer/test/decodeJwt" // 解码JWT

	// out 分组
	PathGetServerList        = "/loginServer/getServerList"       // 获取服务器列表
	PathGetPlayerServerList  = "/loginServer/getPlayerServerList" // 获取玩家服务器列表
	PathClientGetLoginNotice = "/loginServer/getLoginNotice"      // 客户端获取登录公告

	// sgame 分组
	PathTest              = "/loginServer/test"              // 测试接口（GET）
	PathTestPost          = "/loginServer/testPost"          // 测试接口（POST）
	PathReportServerList  = "/loginServer/reportServerList"  // 游戏服上报服务器列表
	PathChangeServerState = "/loginServer/changeServerState" // 游戏服服务器状态变更
	PathSetUserHistory    = "/loginServer/setUserHistory"    // 玩家信息设置
	PathSetUserState      = "/loginServer/setUserState"      // 玩家账号状态设置

	// admin 分组
	// 公告管理接口
	PathCreateLoginNotice      = "/loginServer/loginNotice/create"      // 创建公告
	PathDeleteLoginNotice      = "/loginServer/loginNotice/delete"      // 删除公告
	PathBatchDeleteLoginNotice = "/loginServer/loginNotice/batchDelete" // 批量删除
	PathUpdateLoginNotice      = "/loginServer/loginNotice/update"      // 更新公告
	PathFindLoginNotice        = "/loginServer/loginNotice/find"        // 查询单条 (GET)
	PathGetLoginNoticeList     = "/loginServer/loginNotice/list"        // 获取列表 (GET)
	// IP白名单管理
	PathGetWhitelist      = "/loginServer/whitelist/get"    // 获取指定分组白名单 (GET)
	PathGetAllWhitelists  = "/loginServer/whitelist/getAll" // 获取所有分组白名单 (GET)
	PathSetWhitelist      = "/loginServer/whitelist/set"    // 设置白名单 (POST)
	PathAddWhitelistIP    = "/loginServer/whitelist/add"    // 添加IP (POST)
	PathRemoveWhitelistIP = "/loginServer/whitelist/remove" // 删除IP (POST)
)

// routeCache 路由映射表（path -> Route），在包初始化时构建一次，可供整个包复用。
var routeCache = map[string]Route{
	// out 分组
	PathGetServerList:        {Path: PathGetServerList, Method: MethodGET, Handler: handle_getServerList, IsDebug: false, ApiGroup: ApiGroupOut},
	PathGetPlayerServerList:  {Path: PathGetPlayerServerList, Method: MethodGET, Handler: handle_getPlayerServerList, IsDebug: false, ApiGroup: ApiGroupOut},
	PathClientGetLoginNotice: {Path: PathClientGetLoginNotice, Method: MethodGET, Handler: handle_clientGetLoginNotice, IsDebug: false, ApiGroup: ApiGroupOut},
	// sgame 分组
	PathTest:              {Path: PathTest, Method: MethodGET, Handler: handle_test, IsDebug: true, ApiGroup: ApiGroupSgame},
	PathTestPost:          {Path: PathTestPost, Method: MethodPOST, Handler: handle_testPost, IsDebug: true, ApiGroup: ApiGroupSgame},
	PathReportServerList:  {Path: PathReportServerList, Method: MethodPOST, Handler: handle_reportServerList, IsDebug: false, ApiGroup: ApiGroupSgame},
	PathChangeServerState: {Path: PathChangeServerState, Method: MethodPOST, Handler: handle_changeServerState, IsDebug: false, ApiGroup: ApiGroupSgame},
	PathSetUserHistory:    {Path: PathSetUserHistory, Method: MethodPOST, Handler: handle_SetUserHistory, IsDebug: false, ApiGroup: ApiGroupSgame},
	PathSetUserState:      {Path: PathSetUserState, Method: MethodPOST, Handler: handle_SetUserState, IsDebug: false, ApiGroup: ApiGroupSgame},

	// admin 分组
	// 公告管理接口
	PathCreateLoginNotice:      {Path: PathCreateLoginNotice, Method: MethodPOST, Handler: handle_createLoginNotice, IsDebug: false, ApiGroup: ApiGroupAdminServer},
	PathDeleteLoginNotice:      {Path: PathDeleteLoginNotice, Method: MethodPOST, Handler: handle_deleteLoginNotice, IsDebug: false, ApiGroup: ApiGroupAdminServer},
	PathBatchDeleteLoginNotice: {Path: PathBatchDeleteLoginNotice, Method: MethodPOST, Handler: handle_batchDeleteLoginNotice, IsDebug: false, ApiGroup: ApiGroupAdminServer},
	PathUpdateLoginNotice:      {Path: PathUpdateLoginNotice, Method: MethodPOST, Handler: handle_updateLoginNotice, IsDebug: false, ApiGroup: ApiGroupAdminServer},
	// 查询接口用 GET
	PathFindLoginNotice:    {Path: PathFindLoginNotice, Method: MethodGET, Handler: handle_findLoginNotice, IsDebug: false, ApiGroup: ApiGroupAdminServer},
	PathGetLoginNoticeList: {Path: PathGetLoginNoticeList, Method: MethodGET, Handler: handle_getLoginNoticeList, IsDebug: false, ApiGroup: ApiGroupAdminServer},
	// IP白名单管理接口
	PathGetWhitelist:      {Path: PathGetWhitelist, Method: MethodGET, Handler: handle_getWhitelist, IsDebug: false, ApiGroup: ApiGroupAdminServer},
	PathGetAllWhitelists:  {Path: PathGetAllWhitelists, Method: MethodGET, Handler: handle_getAllWhitelists, IsDebug: false, ApiGroup: ApiGroupAdminServer},
	PathSetWhitelist:      {Path: PathSetWhitelist, Method: MethodPOST, Handler: handle_setWhitelist, IsDebug: false, ApiGroup: ApiGroupAdminServer},
	PathAddWhitelistIP:    {Path: PathAddWhitelistIP, Method: MethodPOST, Handler: handle_addWhitelistIP, IsDebug: false, ApiGroup: ApiGroupAdminServer},
	PathRemoveWhitelistIP: {Path: PathRemoveWhitelistIP, Method: MethodPOST, Handler: handle_removeWhitelistIP, IsDebug: false, ApiGroup: ApiGroupAdminServer},

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
