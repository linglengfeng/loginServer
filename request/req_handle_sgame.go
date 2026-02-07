package request

import (
	"loginServer/src/db"
	"loginServer/src/db/db_mysql"
	"loginServer/src/log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ReportServerReq 上报/注册接口的参数
// 指针类型 (*string, *int) 用于区分 "零值" 和 "未传值(nil)"
type ReportServerReq struct {
	// 必填项：使用 binding:"required" 自动校验
	ClusterID int64 `json:"cluster_id" binding:"required"`
	GameID    int64 `json:"game_id"    binding:"required"`

	// 选填项：没传 JSON 字段时，这些字段会自动为 nil
	Name   *string `json:"name"`
	Addr   *string `json:"addr"`
	Info   *string `json:"info"`
	Port   *int    `json:"port"`
	State  *int    `json:"state"`
	IsShow *int    `json:"is_show"`
	IsNew  *int    `json:"is_new"`
	Desc   *string `json:"desc"`
}

// handle_reportServerList 批量上报/注册
// POST /loginServer/reportServerList
func handle_reportServerList(c *gin.Context) {
	// 1. 定义请求参数切片
	var reqs []ReportServerReq

	// 2. 解析 JSON 数组
	if err := c.ShouldBindJSON(&reqs); err != nil {
		log.Warn("Bind JSON List failed: %v", err)
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "参数错误", nil))
		return
	}

	if len(reqs) == 0 {
		c.JSON(http.StatusOK, retResponse(CodeSuccess, "列表为空", nil))
		return
	}

	// 预分配切片容量，避免 append 时的多次内存分配
	serverModels := make([]db_mysql.GameList, 0, len(reqs))

	for _, req := range reqs {
		// 组装单个 DB 模型
		// 因为 req 中的字段本身就是指针 (*string, *int)，直接赋值即可
		model := db_mysql.GameList{
			ClusterID: req.ClusterID,
			GameID:    req.GameID,
			Name:      req.Name,
			Addr:      req.Addr,
			Info:      req.Info,
			Port:      req.Port,
			State:     req.State,
			IsShow:    req.IsShow,
			IsNew:     req.IsNew,
			Desc:      req.Desc,
		}
		serverModels = append(serverModels, model)
	}
	err := db.BatchUpdateServerInfo(serverModels)
	if err != nil {
		log.Error("handle_reportServerList batch db err: %v", err)
		c.JSON(http.StatusOK, retResponse(CodeError, "批量上报失败", nil))
		return
	}

	//更新缓存
	UpdateCacheServerList(serverModels)
	c.JSON(http.StatusOK, retResponse(CodeSuccess, "批量上报成功", nil))
}

// ChangeStateReq 变更状态接口的参数
type ChangeStateReq struct {
	ClusterID int64 `json:"cluster_id" binding:"required"`
	GameID    int64 `json:"game_id"    binding:"required"`
	State     *int  `json:"state"      binding:"required"`
}

// handle_changeServerState 变更服务器状态
// POST /loginServer/changeServerState
func handle_changeServerState(c *gin.Context) {
	var req []ChangeStateReq

	// 1. 解析 JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "参数错误: 需要 cluster_id, game_id, state", nil))
		return
	}
	serverModels := make([]db_mysql.GameList, 0, len(req))

	for _, req := range req {
		if req.State == nil {
			continue
		}
		model := db_mysql.GameList{
			ClusterID: req.ClusterID,
			GameID:    req.GameID,
			State:     req.State,
		}
		serverModels = append(serverModels, model)

	}

	// 2. 调用 DB
	err := db.UpdateServerState(serverModels)
	if err != nil {
		log.Error("handle_changeServerState db err: %v", err)
		c.JSON(http.StatusOK, retResponse(CodeError, "状态更新失败", nil))
		return
	}

	//更新缓存
	UpdateCacheServerList(serverModels)
	c.JSON(http.StatusOK, retResponse(CodeSuccess, "上报成功", nil))
}

// SetHistoryReq 定义请求参数结构体
// 使用 form 标签支持 PostForm (表单提交)
// 指针类型 (*string, *int) 允许客户端不传这些参数，不传时为 nil
type SetHistoryReq struct {
	AccountID string `form:"account_id" json:"account_id" binding:"required"`
	ClusterID int    `form:"cluster_id" json:"cluster_id" binding:"required"`
	GameID    int    `form:"game_id"    json:"game_id"    binding:"required"`
	PlayerID  int64  `form:"player_id"  json:"player_id"  binding:"required"`
}

// handle_SetUserHistory 玩家登录/升级上报历史记录
// POST /loginServer/setUserHistory
func handle_SetUserHistory(c *gin.Context) {
	var req SetHistoryReq

	// ShouldBind 会根据 Content-Type 自动选择解析方式 (Form-Data 或 JSON)
	// 如果必填参数缺失，err 会不为空
	if err := c.ShouldBind(&req); err != nil {
		log.Warn("handle_SetUserHistory bind params failed: %v", err)
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "参数错误或缺失", nil))
		return
	}

	newItem := db_mysql.PlayerHistoryItem{
		ClusterID: req.ClusterID,
		GameID:    req.GameID,
		PlayerID:  req.PlayerID,
	}

	err := db.SetUserHistory(req.AccountID, newItem)
	if err != nil {
		log.Error("handle_SetUserHistory db err: %v", err)
		c.JSON(http.StatusOK, retResponse(CodeError, "保存历史记录失败", nil))
		return
	}

	c.JSON(http.StatusOK, retResponse(CodeSuccess, "上报成功", nil))
}

// handle_SetUserState 用户账号状态修改
// POST /loginServer/setUserState
func handle_SetUserState(c *gin.Context) {

	accountId, ok1 := c.GetPostForm("account_id")
	stateStr, ok2 := c.GetPostForm("state")

	if !ok1 || !ok2 {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "参数缺失", nil))
		return
	}

	state, err := strconv.Atoi(stateStr)
	if err != nil {
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, "state必须是数字", nil))
		return
	}

	err = db.SetUserState(accountId, state)
	if err != nil {
		log.Error("handle_SetUserState db err: %v", err)
		c.JSON(http.StatusOK, retResponse(CodeError, "更新状态失败", nil))
		return
	}

	c.JSON(http.StatusOK, retResponse(CodeSuccess, "操作成功", nil))
}

// handle_test 测试接口处理函数（GET），能够测试所有的返回情况
// GET /loginServer/test
// 参数说明：
//   - case: 测试场景（可选值：success, error, param_error, all）
//   - with_message: 是否包含消息（true/false，默认 false）
//   - with_data: 是否包含数据（true/false，默认 true）
func handle_test(c *gin.Context) {
	// 使用统一的参数解析函数（支持多种参数传递方式）
	params, _, _ := ParseRequestParams(c)

	// 从合并后的 params 中获取测试参数
	testCase := GetParamString(params, c, "case")
	withMessage := GetParamBool(params, c, "with_message", false)
	withData := GetParamBool(params, c, "with_data", true) // 默认为 true

	// 简单的测试数据
	testData := map[string]any{
		"test_key": "test_value",
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
		c.JSON(http.StatusOK, retResponse(CodeSuccess, message, data))
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
		c.JSON(http.StatusOK, retResponse(CodeError, message, data))
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
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, message, data))
		return

	case "all":
		// 返回所有测试场景说明
		allCases := map[string]any{
			"desc": "测试接口 - 可以测试所有返回情况",
			"params": map[string]string{
				"case":         "测试场景：success(成功)、error(失败)、param_error(参数错误)、all(查看说明)",
				"with_message": "是否包含消息：true/false（默认 false）",
				"with_data":    "是否包含数据：true/false（默认 true）",
			},
		}
		c.JSON(http.StatusOK, retResponse(CodeSuccess, "", allCases))
		return

	default:
		// 默认返回成功响应 + 测试数据（如果没有参数，返回默认测试数据）
		message := ""
		data := any(nil)
		if withMessage {
			message = testMessage
		}
		if withData {
			// 如果有参数，返回参数；否则返回默认测试数据
			if len(params) > 0 {
				data = params
			} else {
				data = testData
			}
		}
		c.JSON(http.StatusOK, retResponse(CodeSuccess, message, data))
	}
}

// handle_testPost 测试接口处理函数（POST），用于接收 Erlang 后端的 POST 请求
// POST /loginServer/testPost
// 参数说明：
//   - case: 测试场景（可选值：success, error, param_error, all）
//   - with_message: 是否包含消息（true/false，默认 false）
//   - with_data: 是否包含数据（true/false，默认 true）
//
// 支持多种参数传递方式（与 Erlang http_game_send 逻辑一致）：
//  1. POST + JSON + params: params 在 body 中，格式为 JSON
//  2. POST + FORM + params: params 在 body 中，格式为 application/x-www-form-urlencoded
//  3. POST + body: 直接使用 body（JSON 格式）
//  4. POST + body + params: body 在 body 中，params 在 URL query string
//  5. POST + 无 body 无 params: 空 body，但需要 Content-Type
func handle_testPost(c *gin.Context) {
	// 使用统一的参数解析函数（支持多种参数传递方式）
	params, _, _ := ParseRequestParams(c)

	// 从合并后的 params 中获取测试参数
	testCase := GetParamString(params, c, "case")
	withMessage := GetParamBool(params, c, "with_message", false)
	withData := GetParamBool(params, c, "with_data", true) // 默认为 true

	// 简单的测试数据
	testData := map[string]any{
		"test_key": "test_value",
	}
	testMessage := "这是测试消息（POST）"

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
		c.JSON(http.StatusOK, retResponse(CodeSuccess, message, data))
		return

	case "error":
		// 失败响应
		message := ""
		data := any(nil)
		if withMessage {
			message = "测试失败消息（POST）"
		}
		if withData {
			data = testData
		}
		c.JSON(http.StatusOK, retResponse(CodeError, message, data))
		return

	case "param_error":
		// 参数错误响应
		message := ""
		data := any(nil)
		if withMessage {
			message = "测试参数错误消息（POST）"
		}
		if withData {
			data = testData
		}
		c.JSON(http.StatusOK, retResponse(CodeBadRequest, message, data))
		return

	case "all":
		// 返回所有测试场景说明
		allCases := map[string]any{
			"desc": "测试接口（POST）- 可以测试所有返回情况",
			"params": map[string]string{
				"case":         "测试场景：success(成功)、error(失败)、param_error(参数错误)、all(查看说明)",
				"with_message": "是否包含消息：true/false（默认 false）",
				"with_data":    "是否包含数据：true/false（默认 true）",
			},
		}
		c.JSON(http.StatusOK, retResponse(CodeSuccess, "", allCases))
		return

	default:
		// 默认返回成功响应 + 测试数据（如果没有参数，返回默认测试数据）
		message := ""
		data := any(nil)
		if withMessage {
			message = testMessage
		}
		if withData {
			// 如果有参数，返回参数；否则返回默认测试数据
			if len(params) > 0 {
				data = params
			} else {
				data = testData
			}
		}
		c.JSON(http.StatusOK, retResponse(CodeSuccess, message, data))
	}
}
