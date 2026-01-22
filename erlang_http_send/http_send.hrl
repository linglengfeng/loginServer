%%%-------------------------------------------------------------------
%%% @author wp
%%% @copyright (C) 2026, <COMPANY>
%%% @doc
%%%
%%% @end
%%% Created : 21. 1月 2026 17:35
%%%-------------------------------------------------------------------
-author("wp").

-define(HTTP_DEFAULT_TIMEOUT, 5000).
-define(HTTP_CONTENT_TYPE_FORM, "application/x-www-form-urlencoded").
-define(HTTP_CONTENT_TYPE_JSON, "application/json").
%% 常用 Content-Type（仅用于对外请求 httpc:request 的 ContentType 字段）
-define(HTTP_CONTENT_TYPE_FORM_UTF8, "application/x-www-form-urlencoded; charset=utf-8").
-define(HTTP_CONTENT_TYPE_JSON_UTF8, "application/json; charset=utf-8").
-define(HTTP_CONTENT_TYPE_TEXT_PLAIN, "text/plain").
-define(HTTP_CONTENT_TYPE_TEXT_PLAIN_UTF8, "text/plain; charset=utf-8").
-define(HTTP_CONTENT_TYPE_TEXT_HTML, "text/html").
-define(HTTP_CONTENT_TYPE_TEXT_HTML_UTF8, "text/html; charset=utf-8").
-define(HTTP_CONTENT_TYPE_XML, "application/xml").
-define(HTTP_CONTENT_TYPE_XML_UTF8, "application/xml; charset=utf-8").
-define(HTTP_CONTENT_TYPE_OCTET_STREAM, "application/octet-stream").
-define(HTTP_CONTENT_TYPE_MULTIPART_FORM, "multipart/form-data").

%% 常用 HTTP Method（统一用宏，避免散落 atom）
-define(HTTP_METHOD_GET, get).
-define(HTTP_METHOD_POST, post).
-define(HTTP_METHOD_PUT, put).
-define(HTTP_METHOD_DELETE, delete).
-define(HTTP_METHOD_HEAD, head).
-define(HTTP_METHOD_PATCH, patch).

%% HTTP 响应结构体（封装层返回的完整响应信息）
-record(http_result_response, {
  vsn = "",              %% HTTP 版本：如 "HTTP/1.1"
  status_code = 0,        %% 状态码：如 200, 404, 500
  reason_phrase = "",     %% 原因短语：如 "OK", "Not Found"
  headers = [],           %% 响应头：[{HeaderName, HeaderValue}]
  body = ""              %% 响应体：binary/string
}).

%% HTTP 请求结构体（用于封装 HTTP 请求的所有信息）
%% 使用流程：
%%   1. 业务层通过 lib_http_send:get_send/1 获取配置模板
%%   2. 调用 lib_http_send:send/2 传入 ID 和透传参数
%%   3. 系统自动执行 send_handle（组装参数）-> 发送 HTTP 请求 -> result_handle（解析返回）
%% 参数位置自动决定：
%%   - GET/HEAD：params 自动放在 URL query string
%%   - POST/PUT/PATCH/DELETE：
%%     * 如果指定了 body，使用 body，params 合并到 URL
%%     * 如果只有 params：
%%       - Content-Type 为 application/x-www-form-urlencoded：params 编码到 body
%%       - Content-Type 为 application/json：params 转换为 JSON 到 body
%%       - 其他：params 放在 URL
-record(http_send, {
  %% ---------------------- 业务层字段（lib_http_send） ----------------------
  id,                    %% 发送唯一ID：?HTTP_SEND_ID_*，用于业务侧路由/查找配置
  map_args = #{},         %% 外部透传参数：业务侧传入（send/2），供 send_handle 使用
  send_handle,            %% 发送前处理：fun(#http_send{}) -> {ok,#http_send{}} | {error,ErrStr}
  result_handle,          %% 返回处理：fun(#http_send{}) -> {ok,Term} | {error,ErrStr}
  result,                 %% 原始返回结果：#http_result_response{}（包含完整响应信息），由封装层写入，供 result_handle 读取

  %% ---------------------- 请求描述字段（封装层 http_game_send） -------------
  path = <<"">>,          %% HTTP 路径：如 <<"/loginServer/test">>（会拼到 URL 后）
  api_group,              %% API 分组：用于选择默认 host/port（例如登录服/游戏服等）
  addr,                   %% 目标地址：可为 host（如 "127.0.0.1"）或完整URL（如 "https://xx"）
  port,                   %% 目标端口：0 表示按 api_group 使用默认端口；>0 表示显式指定
  params = [],            %% Query 参数：[] | proplist | map（通过 cow_qs:qs 编码）  最好直接采用 binary proplist
  headers = [],           %% HTTP 头：[{HeaderName, HeaderValue}]（string/binary 均可）
  body = "",              %% 请求体：binary/string（GET/HEAD 通常为空） json_util:map_to_json
  method = ?HTTP_METHOD_GET, %% 请求方法：get/post/put/delete/head/patch（见宏）
  content_type = ?HTTP_CONTENT_TYPE_JSON, %% Content-Type：用于非 GET/HEAD 的请求体类型
  timeout = ?HTTP_DEFAULT_TIMEOUT, %% 超时（毫秒）：直接传给 httpc:request/4 的 {timeout, Timeout}
  httpc_options = [], %% httpc 选项：[{connect_timeout, 3000}, {autoredirect, true}, {ssl, [...]}, ...]，会合并到 httpc:request/4 的第三个参数
  info = []%% 额外信息
}).

%% api分组
-define(HTTP_API_GROUP_LOGIN, 1).%% 登录服
%% 登录服统一相应 body #{"data" => #{},"message" => <<"success">>,"status" => 0}
-define(HTTP_API_GROUP_LOGIN_CODE_SUCCESS, 0).%% 成功
-define(HTTP_API_GROUP_LOGIN_CODE_FAIL, 1001).%%一般错误
-define(HTTP_API_GROUP_LOGIN_CODE_PARAMS_FAIL, 1002).%%参数错误


%% 定义发送唯一id, 每种占用一百
%% 100 登录服
-define(HTTP_SEND_ID_LOGIN_TEST, 100).%% 登录服 测试
-define(HTTP_SEND_ID_LOGIN_TEST_POST, 101).%% 登录服 测试


