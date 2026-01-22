package request

import (
	"encoding/json"
	"io"
	"loginServer/src/log"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ParseBody 根据请求的 Content-Type 自动解析请求体
// 支持 JSON (application/json) 和表单 (application/x-www-form-urlencoded, multipart/form-data)
// 如果 Content-Type 未指定或无法确定，优先尝试 JSON 解析，失败后尝试表单解析
func ParseBody(c *gin.Context) (map[string]any, error) {
	contentType := c.GetHeader("Content-Type")
	contentType = strings.ToLower(strings.TrimSpace(contentType))

	// 移除可能的字符集信息（如 application/json; charset=utf-8）
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}

	// 根据 Content-Type 选择解析方式
	switch contentType {
	case "application/json":
		return parseJSONBody(c)
	case "application/x-www-form-urlencoded", "multipart/form-data":
		return parseFormBody(c)
	default:
		// Content-Type 未指定或无法确定时，先尝试 JSON，失败后尝试表单
		if params, err := parseJSONBody(c); err == nil {
			return params, nil
		}
		return parseFormBody(c)
	}
}

// parseJSONBody 解析 JSON 请求体
// 支持空 body（返回空 map，不返回错误）
func parseJSONBody(c *gin.Context) (map[string]any, error) {
	// 检查 body 是否为空
	if c.Request.Body == nil {
		return make(map[string]any), nil
	}

	// 读取 body 内容
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	// 恢复 body，以便后续使用
	c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	// 如果 body 为空，返回空 map
	if len(bodyBytes) == 0 {
		return make(map[string]any), nil
	}

	// 直接解析 JSON（不使用 ShouldBindJSON，因为已经读取了 body）
	var params map[string]any
	if err := json.Unmarshal(bodyBytes, &params); err != nil {
		return nil, err
	}
	if params == nil {
		params = make(map[string]any)
	}
	return params, nil
}

// parseFormBody 解析表单请求体
// 支持 POST/PUT/PATCH 的 body 和 GET 请求的 body
func parseFormBody(c *gin.Context) (map[string]any, error) {
	// 先尝试 ParseForm（适用于 POST/PUT/PATCH 请求）
	if err := c.Request.ParseForm(); err != nil {
		return nil, err
	}

	params := make(map[string]any)

	// 优先使用 PostForm（POST/PUT/PATCH 请求的 body 数据）
	if len(c.Request.PostForm) > 0 {
		for key, values := range c.Request.PostForm {
			if len(values) > 0 {
				params[key] = values[0]
			}
		}
		return params, nil
	}

	// 对于 GET 请求，需要直接从 body 读取
	// 因为 ParseForm() 对于 GET 请求只会解析查询参数，不会解析 body
	if c.Request.Method == "GET" && c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err == nil && len(bodyBytes) > 0 {
			// 恢复 body，以便后续使用
			c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

			// 解析 urlencoded 格式的 body
			values, err := url.ParseQuery(string(bodyBytes))
			if err == nil {
				for key, valSlice := range values {
					if len(valSlice) > 0 {
						params[key] = valSlice[0]
					}
				}
				if len(params) > 0 {
					return params, nil
				}
			}
		}
	}

	// 如果都没有，尝试从 Form 获取（可能包含查询参数）
	if len(c.Request.Form) > 0 {
		for key, values := range c.Request.Form {
			if len(values) > 0 {
				params[key] = values[0]
			}
		}
	}

	return params, nil
}

// JsonBody 解析请求体中的 JSON 数据（已废弃，建议使用 ParseBody）
// 保留此函数以保持向后兼容
func JsonBody(c *gin.Context) (map[string]any, error) {
	return parseJSONBody(c)
}

// FormBody 解析请求中的表单数据（已废弃，建议使用 ParseBody）
// 保留此函数以保持向后兼容
func FormBody(c *gin.Context) (map[string]any, error) {
	return parseFormBody(c)
}

// ParseRequestParams 解析请求参数（支持多种参数传递方式，与 Erlang http_game_send 逻辑一致）
// 支持以下场景：
//  1. POST + JSON + params: params 在 body 中，格式为 JSON
//  2. POST + FORM + params: params 在 body 中，格式为 application/x-www-form-urlencoded
//  3. POST + body: 直接使用 body（JSON 格式）
//  4. POST + body + params: body 在 body 中，params 在 URL query string
//  5. POST + 无 body 无 params: 空 body，但需要 Content-Type
//
// 返回：
//   - mergedParams: 合并后的参数（query 参数优先级更高，会覆盖 body 中的同名参数）
//   - rawBody: 原始 body 内容（用于调试）
//   - debugInfo: 调试信息（包含 body_params、query_params、merged_params、raw_body、content_type）
func ParseRequestParams(c *gin.Context) (mergedParams map[string]any, rawBody string, debugInfo map[string]any) {
	// 读取 body 内容（只读取一次，避免重复读取）
	var bodyBytes []byte
	if c.Request.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(c.Request.Body)
		if err == nil {
			rawBody = string(bodyBytes)
			// 恢复 body，以便后续使用
			c.Request.Body = io.NopCloser(strings.NewReader(rawBody))
		}
	}

	// 解析请求体（根据 Content-Type 自动选择解析方式）
	bodyParams := make(map[string]any)
	contentType := c.GetHeader("Content-Type")
	contentType = strings.ToLower(strings.TrimSpace(contentType))

	// 移除可能的字符集信息
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}

	// 根据 Content-Type 解析 body
	if len(bodyBytes) > 0 {
		switch contentType {
		case "application/json":
			// 解析 JSON
			if err := json.Unmarshal(bodyBytes, &bodyParams); err == nil {
				if bodyParams == nil {
					bodyParams = make(map[string]any)
				}
			}
		case "application/x-www-form-urlencoded", "multipart/form-data":
			// 解析表单
			if c.Request.Method == "GET" {
				// GET 请求：直接从 bodyBytes 解析（因为 ParseForm 不会解析 GET 请求的 body）
				values, err := url.ParseQuery(string(bodyBytes))
				if err == nil {
					for key, valSlice := range values {
						if len(valSlice) > 0 {
							bodyParams[key] = valSlice[0]
						}
					}
				}
			} else {
				// POST/PUT/PATCH 等请求：使用 ParseForm
				if err := c.Request.ParseForm(); err == nil {
					for key, values := range c.Request.PostForm {
						if len(values) > 0 {
							bodyParams[key] = values[0]
						}
					}
				}
			}
		default:
			// 未指定 Content-Type 时，先尝试 JSON，失败后尝试表单
			if err := json.Unmarshal(bodyBytes, &bodyParams); err != nil {
				// JSON 解析失败，尝试表单
				if c.Request.Method == "GET" {
					// GET 请求：直接从 bodyBytes 解析
					values, err := url.ParseQuery(string(bodyBytes))
					if err == nil {
						for key, valSlice := range values {
							if len(valSlice) > 0 {
								bodyParams[key] = valSlice[0]
							}
						}
					}
				} else {
					// POST/PUT/PATCH 等请求：使用 ParseForm
					if err := c.Request.ParseForm(); err == nil {
						for key, values := range c.Request.PostForm {
							if len(values) > 0 {
								bodyParams[key] = values[0]
							}
						}
					}
				}
			} else if bodyParams == nil {
				bodyParams = make(map[string]any)
			}
		}
	}

	// 合并 body 参数和 query 参数（query 参数优先级更高）
	mergedParams = make(map[string]any)
	// 先添加 body 参数
	for k, v := range bodyParams {
		mergedParams[k] = v
	}
	// 再添加 query 参数（会覆盖 body 中的同名参数）
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			mergedParams[k] = v[0]
		}
	}

	// 构建调试信息
	debugInfo = map[string]any{
		"body_params":   bodyParams,
		"query_params":  c.Request.URL.Query(),
		"merged_params": mergedParams,
		"raw_body":      rawBody,
		"content_type":  c.GetHeader("Content-Type"),
	}

	// 打印解析出来的参数（一行输出）
	log.Info("ParseRequestParams - method: %s, path: %s, content_type: %s, raw_body: %s, body_params: %v, query_params: %v, merged_params: %v",
		c.Request.Method,
		c.Request.URL.Path,
		c.GetHeader("Content-Type"),
		rawBody,
		bodyParams,
		c.Request.URL.Query(),
		mergedParams,
	)

	return mergedParams, rawBody, debugInfo
}

// GetParamString 从 params map 中获取字符串参数，如果不存在则从查询参数获取
func GetParamString(params map[string]any, c *gin.Context, key string) string {
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return c.Query(key)
}

// GetParamBool 从 params map 中获取布尔参数，如果不存在则从查询参数获取
// 支持字符串 "true"/"false" 和布尔类型
func GetParamBool(params map[string]any, c *gin.Context, key string, defaultValue bool) bool {
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			return str == "true"
		} else if b, ok := val.(bool); ok {
			return b
		}
	}
	// 从查询参数获取
	queryVal := c.Query(key)
	if queryVal == "" {
		return defaultValue
	}
	return queryVal == "true"
}

// GetParamInt 从 params map 中获取整数参数，如果不存在则从查询参数获取
// 支持字符串数字和整数类型
func GetParamInt(params map[string]any, c *gin.Context, key string, defaultValue int) int {
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			if i, err := strconv.Atoi(str); err == nil {
				return i
			}
		} else if i, ok := val.(int); ok {
			return i
		} else if i, ok := val.(int64); ok {
			return int(i)
		} else if i, ok := val.(int32); ok {
			return int(i)
		} else if f, ok := val.(float64); ok {
			return int(f)
		}
	}
	// 从查询参数获取
	queryVal := c.Query(key)
	if queryVal == "" {
		return defaultValue
	}
	if i, err := strconv.Atoi(queryVal); err == nil {
		return i
	}
	return defaultValue
}

// GetParamInt64 从 params map 中获取 int64 参数，如果不存在则从查询参数获取
// 支持字符串数字和整数类型
func GetParamInt64(params map[string]any, c *gin.Context, key string, defaultValue int64) int64 {
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			if i, err := strconv.ParseInt(str, 10, 64); err == nil {
				return i
			}
		} else if i, ok := val.(int64); ok {
			return i
		} else if i, ok := val.(int); ok {
			return int64(i)
		} else if i, ok := val.(int32); ok {
			return int64(i)
		} else if f, ok := val.(float64); ok {
			return int64(f)
		}
	}
	// 从查询参数获取
	queryVal := c.Query(key)
	if queryVal == "" {
		return defaultValue
	}
	if i, err := strconv.ParseInt(queryVal, 10, 64); err == nil {
		return i
	}
	return defaultValue
}

// GetParamFloat64 从 params map 中获取 float64 参数，如果不存在则从查询参数获取
// 支持字符串数字和浮点数类型
func GetParamFloat64(params map[string]any, c *gin.Context, key string, defaultValue float64) float64 {
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			if f, err := strconv.ParseFloat(str, 64); err == nil {
				return f
			}
		} else if f, ok := val.(float64); ok {
			return f
		} else if f, ok := val.(float32); ok {
			return float64(f)
		} else if i, ok := val.(int); ok {
			return float64(i)
		} else if i, ok := val.(int64); ok {
			return float64(i)
		}
	}
	// 从查询参数获取
	queryVal := c.Query(key)
	if queryVal == "" {
		return defaultValue
	}
	if f, err := strconv.ParseFloat(queryVal, 64); err == nil {
		return f
	}
	return defaultValue
}

// GetParamFloat32 从 params map 中获取 float32 参数，如果不存在则从查询参数获取
// 支持字符串数字和浮点数类型
func GetParamFloat32(params map[string]any, c *gin.Context, key string, defaultValue float32) float32 {
	return float32(GetParamFloat64(params, c, key, float64(defaultValue)))
}
