package request

import (
	"loginServer/src/db"
	"loginServer/src/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// handle_getServerList 获取服务器列表处理函数
// GET /loginServer/getServerList
func handle_getServerList(c *gin.Context) {
	servers, err := GetServerList()
	if err != nil {
		log.Error("handle_getServerList failed, err: %v", err)
		c.JSON(http.StatusOK, retResponse(CodeError, "获取服务器列表失败", nil))
		return
	}
	c.JSON(http.StatusOK, retResponse(CodeSuccess, "", servers))
}

// handle_getPlayerServerList 获取玩家服务器列表处理函数
// GET /loginServer/getPlayerServerList
func handle_getPlayerServerList(c *gin.Context) {
	accountId, ok1 := c.GetQuery("account_id")

	if !ok1 {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "缺少必填参数: account_id", nil))
		return
	}
	// 如果获取不到，说明没登录或者中间件有问题
	if accountId == "" {
		c.JSON(http.StatusOK, retResponse(CodeError, "请求参数错误", nil))
		return
	}

	// 传入 accountID
	playerHistory, err := db.GetUserHistory(accountId)
	if err != nil {
		log.Error("handle_getPlayerServerList failed, err: %v", err)
		c.JSON(http.StatusOK, retResponse(CodeError, "获取服务器列表失败", nil))
		return
	}

	c.JSON(http.StatusOK, retResponse(CodeSuccess, "", playerHistory))
}

// handle_clientGetLoginNotice 客户端(游戏)获取公告
// GET /loginServer/getLoginNotice
func handle_clientGetLoginNotice(c *gin.Context) {
	// 直接走内存缓存，无需查库，高性能
	list, _ := GetLoginNotice()

	// 返回给客户端的数据结构，根据你的客户端协议定义
	c.JSON(http.StatusOK, retResponse(CodeSuccess, "获取成功", list))
}
