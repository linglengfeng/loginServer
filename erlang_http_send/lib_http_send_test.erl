%%%-------------------------------------------------------------------
%%% @author wp
%%% @copyright (C) 2026, <COMPANY>
%%% @doc
%%%
%%% @end
%%% Created : 22. 1月 2026 14:01
%%%-------------------------------------------------------------------
-module(lib_http_send_test).
-author("wp").

%% API
-compile([nowarn_export_all, export_all, nowarn_unused_vars]). %% 导出所有且不警告

-include("http_send.hrl").

test() ->
  %% ====================================================================
  %% 测试1：通过业务层接口调用（推荐方式）
  %% ====================================================================
  io:format("=== 测试1：业务层接口调用 ===~n"),
  Result1_1 = lib_http_send:send(100, #{}),  %% GET 请求，params 在 send_handle 中设置
  io:format("测试1-1 返回结果: ~p~n", [Result1_1]),
  Result1_2 = lib_http_send:send(101, #{}),  %% POST 请求，无 params 无 body
  io:format("测试1-2 返回结果: ~p~n", [Result1_2]),

  %% ====================================================================
  %% 测试2：GET 请求 - params 自动放在 URL
  %% 参数说明：
  %%   - case: 测试场景（可选值：success, error, param_error, all）
  %%   - with_message: 是否包含消息（true/false，默认 false）
  %%   - with_data: 是否包含数据（true/false，默认 true）
  %% ====================================================================
  io:format("=== 测试2：GET 请求，params 在 URL ===~n"),
  Result2 = http_game_send:send(#http_send{
    id = 100, map_args = #{},
    send_handle = fun lib_http_send:send_handle/1,
    result_handle = fun lib_http_send:result_handle/1,
    path = <<"/loginServer/test">>,
    api_group = 1,
    params = [{<<"case">>, <<"all">>}, {<<"with_message">>, <<"true">>}, {<<"with_data">>, <<"true">>}],
    method = get,
    content_type = "application/json"
  }),
  io:format("测试2 返回结果: ~p~n", [Result2]),

  %% ====================================================================
  %% 测试3：POST + JSON + params - params 自动转成 JSON 到 body
  %% 参数说明：
  %%   - case: 测试场景（可选值：success, error, param_error, all）
  %%   - with_message: 是否包含消息（true/false，默认 false）
  %%   - with_data: 是否包含数据（true/false，默认 true）
  %% ====================================================================
  io:format("=== 测试3：POST + JSON + params，params 转 JSON 到 body ===~n"),
  Result3 = http_game_send:send(#http_send{
    id = 101, map_args = #{},
    send_handle = fun lib_http_send:send_handle/1,
    result_handle = fun lib_http_send:result_handle/1,
    path = <<"/loginServer/testPost">>,
    api_group = 1,
    params = [{<<"case">>, <<"success">>}, {<<"with_message">>, <<"false">>}, {<<"with_data">>, <<"true">>}],
    method = post,
    content_type = "application/json"
  }),
  io:format("测试3 返回结果: ~p~n", [Result3]),

  %% ====================================================================
  %% 测试4：POST + FORM + params - params 自动编码到 body
  %% 参数说明：
  %%   - case: 测试场景（可选值：success, error, param_error, all）
  %%   - with_message: 是否包含消息（true/false，默认 false）
  %%   - with_data: 是否包含数据（true/false，默认 true）
  %% ====================================================================
  io:format("=== 测试4：POST + FORM + params，params 编码到 body ===~n"),
  Result4 = http_game_send:send(#http_send{
    id = 101, map_args = #{},
    send_handle = fun lib_http_send:send_handle/1,
    result_handle = fun lib_http_send:result_handle/1,
    path = <<"/loginServer/testPost">>,
    api_group = 1,
    params = [{<<"case">>, <<"error">>}, {<<"with_message">>, <<"true">>}, {<<"with_data">>, <<"false">>}],
    method = post,
    content_type = "application/x-www-form-urlencoded"
  }),
  io:format("测试4 返回结果: ~p~n", [Result4]),

  %% ====================================================================
  %% 测试5：POST + body - 直接使用 body
  %% 参数说明：
  %%   - case: 测试场景（可选值：success, error, param_error, all）
  %%   - with_message: 是否包含消息（true/false，默认 false）
  %%   - with_data: 是否包含数据（true/false，默认 true）
  %% ====================================================================
  io:format("=== 测试5：POST + body，直接使用 body ===~n"),
  Result5 = http_game_send:send(#http_send{
    id = 101, map_args = #{},
    send_handle = fun lib_http_send:send_handle/1,
    result_handle = fun lib_http_send:result_handle/1,
    path = <<"/loginServer/testPost">>,
    api_group = 1,
    body = json_util:map_to_json(#{<<"case">> => <<"all">>, <<"with_message">> => <<"true">>, <<"with_data">> => <<"true">>}),
    method = post,
    content_type = "application/json"
  }),
  io:format("测试5 返回结果: ~p~n", [Result5]),

  %% ====================================================================
  %% 测试6：POST + body + params - body 优先，params 合并到 URL
  %% 参数说明：
  %%   - case: 测试场景（可选值：success, error, param_error, all）
  %%   - with_message: 是否包含消息（true/false，默认 false）
  %%   - with_data: 是否包含数据（true/false，默认 true）
  %% ====================================================================
  io:format("=== 测试6：POST + body + params，body 优先，params 在 URL ===~n"),
  Result6 = http_game_send:send(#http_send{
    id = 101, map_args = #{},
    send_handle = fun lib_http_send:send_handle/1,
    result_handle = fun lib_http_send:result_handle/1,
    path = <<"/loginServer/testPost">>,
    api_group = 1,
    body = json_util:map_to_json(#{<<"case">> => <<"error">>, <<"with_message">> => <<"false">>, <<"with_data">> => <<"false">>}),
    params = [{<<"case">>, <<"param_error">>}, {<<"with_message">>, <<"true">>}, {<<"with_data">>, <<"true">>}],
    method = post,
    content_type = "application/json"
  }),
  io:format("测试6 返回结果: ~p~n", [Result6]),

  %% ====================================================================
  %% 测试7：POST + 无 body 无 params - 使用空的 body（确保有 Content-Type）
  %% ====================================================================
  io:format("=== 测试7：POST + 无 body 无 params，使用空 body ===~n"),
  Result7 = http_game_send:send(#http_send{
    id = 101, map_args = #{},
    send_handle = fun lib_http_send:send_handle/1,
    result_handle = fun lib_http_send:result_handle/1,
    path = <<"/loginServer/testPost">>,
    api_group = 1,
    params = [],
    body = "",
    method = post,
    content_type = "application/json"
  }),
  io:format("测试7 返回结果: ~p~n", [Result7]),

  %% ====================================================================
  %% 测试8：DELETE - 可以没有 body
  %% 参数说明：
  %%   - case: 测试场景（可选值：success, error, param_error, all）
  %%   - with_message: 是否包含消息（true/false，默认 false）
  %%   - with_data: 是否包含数据（true/false，默认 true）
  %% ====================================================================
  io:format("=== 测试8：DELETE 请求，可以没有 body ===~n"),
  Result8 = http_game_send:send(#http_send{
    id = 101, map_args = #{},
    send_handle = fun lib_http_send:send_handle/1,
    result_handle = fun lib_http_send:result_handle/1,
    path = <<"/loginServer/testDelete">>,
    api_group = 1,
    params = [{<<"case">>, <<"success">>}, {<<"with_message">>, <<"false">>}, {<<"with_data">>, <<"false">>}],
    method = delete,
    content_type = "application/json"
  }),
  io:format("测试8 返回结果: ~p~n", [Result8]),

  %% ====================================================================
  %% 测试9：PUT + JSON + params
  %% 参数说明：
  %%   - case: 测试场景（可选值：success, error, param_error, all）
  %%   - with_message: 是否包含消息（true/false，默认 false）
  %%   - with_data: 是否包含数据（true/false，默认 true）
  %% ====================================================================
  io:format("=== 测试9：PUT + JSON + params ===~n"),
  Result9 = http_game_send:send(#http_send{
    id = 101, map_args = #{},
    send_handle = fun lib_http_send:send_handle/1,
    result_handle = fun lib_http_send:result_handle/1,
    path = <<"/loginServer/testPut">>,
    api_group = 1,
    params = [{<<"case">>, <<"all">>}, {<<"with_message">>, <<"true">>}, {<<"with_data">>, <<"true">>}],
    method = put,
    content_type = "application/json"
  }),
  io:format("测试9 返回结果: ~p~n", [Result9]),

  io:format("=== 所有测试完成 ===~n"),
  ok.
