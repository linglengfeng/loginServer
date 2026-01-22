%%%-------------------------------------------------------------------
%%% @author wp
%%% @copyright (C) 2026, <COMPANY>
%%% @doc
%%%
%%% @end
%%% Created : 21. 1月 2026 17:21
%%%-------------------------------------------------------------------
-module(lib_http_send).
-author("wp").

%% API
-export([
  send/2,%% 根据发送ID和透传参数触发一次HTTP请求
  get_send/1%% 获取相关定义
]).

%% 默认的发送前/发送后处理函数（可在配置中替换为业务自己的实现）
-export([
  send_handle/1,%% 发送前自定义组装与校验
  result_handle/1,%% 发送后自定义结果解析
  get_group_host_port/1%% 根据分组&集群获取 host/port
]).

-include("http_send.hrl").

%% ====================================================================
%% 业务配置：根据发送ID返回对应的 #http_send 模板
%% ====================================================================
%% 登录服相关
%% 登录服：测试
get_send(Id = ?HTTP_SEND_ID_LOGIN_TEST) ->
  #http_send{id = Id, path = <<"/loginServer/test">>, method = ?HTTP_METHOD_GET, send_handle = fun lib_http_send:send_handle/1, result_handle = fun lib_http_send:result_handle/1, api_group = ?HTTP_API_GROUP_LOGIN};
get_send(Id = ?HTTP_SEND_ID_LOGIN_TEST_POST) ->
  #http_send{id = Id, path = <<"/loginServer/testPost">>, method = ?HTTP_METHOD_POST, send_handle = fun lib_http_send:send_handle/1, result_handle = fun lib_http_send:result_handle/1, api_group = ?HTTP_API_GROUP_LOGIN};

get_send(_) ->
  undefined.

%% -------------------------------------------------------------------
%% 对外统一调用入口
%% Id:    ?HTTP_SEND_ID_*
%% MapArgs: 外部透传参数（会挂到 #http_send.map_args）
%% 返回:
%%   {ok, Term} | {error, ErrStr}
%% -------------------------------------------------------------------
send(Id, MapArgs) ->
  case lib_http_send:get_send(Id) of
    #http_send{} = HttpSend ->
      HttpSend1 = HttpSend#http_send{map_args = MapArgs},
      http_game_send:send(HttpSend1);
    _ ->
      {error, "http_send_error_not_find"}
  end.

%% ====================================================================
%% 默认处理函数：发送前组装参数
%% ====================================================================
%% {ok, #http_send{}} | {error, ErrStr}
send_handle(HttpSend = #http_send{id = ?HTTP_SEND_ID_LOGIN_TEST}) ->
  %% 登录服测试接口：组装 query 参数
  Params = [
    {<<"case">>, <<"all">>},
    {<<"with_message">>, <<"true">>},
    {<<"with_data">>, <<"true">>}
  ],
  {ok, HttpSend#http_send{params = Params}};
send_handle(HttpSend = #http_send{}) ->
  {ok, HttpSend}.

%% {ok, Term} | {error, ErrStr}
result_handle(_HttpSend = #http_send{api_group = ?HTTP_API_GROUP_LOGIN, result = _Result = #http_result_response{body = Body}}) ->
  {ok, json_util:json_to_map(Body)}.

%% ====================================================================
%% 业务: 根据 api_group & cluster 选择 host/port
%% ====================================================================
%% 返回: {Host, Port}
%% - ?HTTP_API_GROUP_LOGIN: 登录服 host/port
%% - 其它: 默认走游戏服 http host/port
%% 若 Port0 > 0，则优先使用 Port0 作为端口
get_group_host_port(#http_send{api_group = ?HTTP_API_GROUP_LOGIN, port = Port0}) ->
  Cluster = cluster:get_cluster(),
  Host = db_cluster:get_cluster_login_host(Cluster),
  Port = case is_integer(Port0) andalso Port0 > 0 of
           true -> Port0;
           _ -> db_cluster:get_cluster_login_port(Cluster)
         end,
  {Host, Port}.






