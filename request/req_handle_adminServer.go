package request

import (
	"loginServer/src/db"
	"loginServer/src/db/db_mysql"
	"loginServer/src/log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 批量 ID 请求
type IDsReq struct {
	IDs []uint64 `json:"ids"`
}

// 列表查询请求
type PageReq struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"pageSize"`
	Title      string `form:"title"`
	NoticeType int    `form:"notice_type"`
	IsEnable   *int   `form:"is_enable"` // 指针类型以区分0和空
}

// handle_createLoginNotice 创建公告
// POST /loginNotice/create
func handle_createLoginNotice(c *gin.Context) {
	var notice db_mysql.LoginNotice
	if err := c.ShouldBindJSON(&notice); err != nil {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "参数错误", nil))
		return
	}

	if err := db.CreateLoginNotice(notice); err != nil {
		log.Error("CreateLoginNotice db err: %v", err)
		c.JSON(http.StatusOK, retResponse(CodeError, "创建失败", nil))
		return
	}

	// 自动刷新缓存
	UpdateNoticeList()
	c.JSON(http.StatusOK, retResponse(CodeSuccess, "创建成功", nil))
}

// handle_deleteLoginNotice 删除公告
// POST /loginNotice/delete (GVA 转发过来的是 POST JSON Body)
func handle_deleteLoginNotice(c *gin.Context) {
	var Id uint64

	params, _, _ := ParseRequestParams(c)

	IdStr := GetParamString(params, c, "id")

	if IdStr == "" {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "id不能为空", nil))
		return
	}

	if idVal, err := strconv.ParseUint(IdStr, 10, 64); err == nil {
		Id = idVal
	}

	if err := db.DeleteLoginNotice(Id); err != nil {
		c.JSON(http.StatusOK, retResponse(CodeError, "删除失败", nil))
		return
	}

	UpdateNoticeList()
	c.JSON(http.StatusOK, retResponse(CodeSuccess, "删除成功", nil))
}

// handle_batchDeleteLoginNotice 批量删除
// POST /loginNotice/batchDelete
func handle_batchDeleteLoginNotice(c *gin.Context) {
	var req IDsReq
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "参数错误", nil))
		return
	}

	if err := db.BatchDeleteLoginNotice(req.IDs); err != nil {
		c.JSON(http.StatusOK, retResponse(CodeError, "批量删除失败", nil))
		return
	}

	UpdateNoticeList()
	c.JSON(http.StatusOK, retResponse(CodeSuccess, "批量删除成功", nil))
}

// handle_updateLoginNotice 更新公告
// POST /loginNotice/update
func handle_updateLoginNotice(c *gin.Context) {
	var notice db_mysql.LoginNotice
	if err := c.ShouldBindJSON(&notice); err != nil {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "参数错误", nil))
		return
	}

	if err := db.UpdateLoginNotice(notice); err != nil {
		c.JSON(http.StatusOK, retResponse(CodeError, "更新失败", nil))
		return
	}

	UpdateNoticeList()
	c.JSON(http.StatusOK, retResponse(CodeSuccess, "更新成功", nil))
}

// handle_findLoginNotice 查询单条
// GET /loginNotice/find?id=xx
func handle_findLoginNotice(c *gin.Context) {
	idStr := c.Query("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)
	if id == 0 {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "ID无效", nil))
		return
	}

	notice, err := db.FindLoginNotice(id)
	if err != nil {
		c.JSON(http.StatusOK, retResponse(CodeError, "查询失败", nil))
		return
	}

	c.JSON(http.StatusOK, retResponse(CodeSuccess, "查询成功", notice))
}

// handle_getLoginNoticeList 获取列表 (给后台用)
// GET /loginNotice/list
func handle_getLoginNoticeList(c *gin.Context) {
	var req PageReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "参数错误", nil))
		return
	}

	// 默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	list, total, err := db.GetLoginNoticeList(req.Page, req.PageSize, req.Title, req.NoticeType, req.IsEnable)
	if err != nil {
		c.JSON(http.StatusOK, retResponse(CodeError, "获取列表失败", nil))
		return
	}

	c.JSON(http.StatusOK, retResponse(CodeSuccess, "获取成功", gin.H{
		"list":     list,
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
	}))
}

// ========== IP白名单管理接口 ==========

// WhitelistSetReq 设置白名单请求
type WhitelistSetReq struct {
	ApiGroup string   `json:"api_group" binding:"required"` // API分组：sgame, adminServer, out, test
	IPs      []string `json:"ips"`                          // IP列表（支持CIDR格式）
}

// WhitelistAddReq 添加IP请求
type WhitelistAddReq struct {
	ApiGroup string `json:"api_group" binding:"required"` // API分组
	IP       string `json:"ip" binding:"required"`         // 要添加的IP（支持CIDR格式）
}

// WhitelistRemoveReq 删除IP请求
type WhitelistRemoveReq struct {
	ApiGroup string `json:"api_group" binding:"required"` // API分组
	IP       string `json:"ip" binding:"required"`         // 要删除的IP
}

// handle_getWhitelist 获取指定分组的白名单
// GET /whitelist/get?api_group=sgame
func handle_getWhitelist(c *gin.Context) {
	apiGroup := c.Query("api_group")
	if apiGroup == "" {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "参数错误: api_group不能为空", nil))
		return
	}

	ips := GetWhitelist(apiGroup)
	c.JSON(http.StatusOK, retResponse(CodeSuccess, "", map[string]interface{}{
		"api_group": apiGroup,
		"ips":       ips,
	}))
}

// handle_getAllWhitelists 获取所有分组的白名单
// GET /whitelist/getAll
func handle_getAllWhitelists(c *gin.Context) {
	allWhitelists := GetAllWhitelists()
	c.JSON(http.StatusOK, retResponse(CodeSuccess, "", allWhitelists))
}

// handle_setWhitelist 设置指定分组的白名单（完全替换）
// POST /whitelist/set
func handle_setWhitelist(c *gin.Context) {
	var req WhitelistSetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "参数错误: "+err.Error(), nil))
		return
	}

	if err := SetWhitelist(req.ApiGroup, req.IPs); err != nil {
		log.Error("SetWhitelist failed: %v", err)
		c.JSON(http.StatusOK, retResponse(CodeError, "设置失败: "+err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, retResponse(CodeSuccess, "设置成功", nil))
}

// handle_addWhitelistIP 向指定分组添加IP
// POST /whitelist/add
func handle_addWhitelistIP(c *gin.Context) {
	var req WhitelistAddReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "参数错误: "+err.Error(), nil))
		return
	}

	if err := AddIP(req.ApiGroup, req.IP); err != nil {
		log.Error("AddIP failed: %v", err)
		c.JSON(http.StatusOK, retResponse(CodeError, "添加失败: "+err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, retResponse(CodeSuccess, "添加成功", nil))
}

// handle_removeWhitelistIP 从指定分组删除IP
// POST /whitelist/remove
func handle_removeWhitelistIP(c *gin.Context) {
	var req WhitelistRemoveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "参数错误: "+err.Error(), nil))
		return
	}

	if err := RemoveIP(req.ApiGroup, req.IP); err != nil {
		log.Error("RemoveIP failed: %v", err)
		c.JSON(http.StatusOK, retResponse(CodeError, "删除失败: "+err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, retResponse(CodeSuccess, "删除成功", nil))
}
