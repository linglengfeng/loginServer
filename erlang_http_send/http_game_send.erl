%%%-------------------------------------------------------------------
%%% @author wp
%%% @copyright (C) 2026, <COMPANY>
%%% @doc
%%%
%%% @end
%%% Created : 21. 1月 2026 17:06
%%%-------------------------------------------------------------------
-module(http_game_send).
-author("wp").

%% API
-export([
  send/1
]).

-include("http_send.hrl").

%% ====================================================================
%% API
%% ====================================================================
%% send/1:
%% - 第一步：执行自定义 send_handle/1（组装/校验参数）
%% - 第二步：根据 #http_send 内部字段发起 HTTP 请求（httpc）
%% - 第三步：执行自定义 result_handle/1（解析返回）
%% 返回:
%%   {ok, Term} | {error, ErrStr}
send(HttpSend = #http_send{}) ->
  case run_send_handle(HttpSend) of
    {ok, HttpSend1} ->
      case do_send_http(HttpSend1) of
        {ok, RespData} ->
          run_result_handle(HttpSend1#http_send{result = RespData});
        {error, ErrStr} ->
          {error, ErrStr}
      end;
    {error, ErrStr} ->
      {error, ErrStr}
  end;
send(_) ->
  {error, "http_send_error_params"}.

%% ====================================================================
%% Internal: send_handle / result_handle 调度
%% ====================================================================
run_send_handle(HttpSend = #http_send{send_handle = undefined}) ->
  {ok, HttpSend};
run_send_handle(HttpSend = #http_send{send_handle = SendHandle}) when is_function(SendHandle, 1) ->
  case SendHandle(HttpSend) of
    {ok, #http_send{} = NewSend} ->
      {ok, NewSend};
    {error, ErrStr} ->
      {error, ErrStr};
    Other ->
      {error, util:term_to_string({http_send_error_send_handle_invalid_return, Other})}
  end;
run_send_handle(Other) ->
  {error, util:term_to_string({http_send_error_send_handle_invalid, Other})}.

run_result_handle(#http_send{result_handle = undefined, result = Result}) ->
  {ok, Result};
run_result_handle(HttpSend = #http_send{result_handle = ResultHandle}) when is_function(ResultHandle, 1) ->
  case ResultHandle(HttpSend) of
    {ok, Term} ->
      {ok, Term};
    {error, ErrStr} ->
      {error, ErrStr};
    Other ->
      {error, util:term_to_string({http_send_error_result_handle_invalid_return, Other})}
  end;
run_result_handle(Other) ->
  {error, util:term_to_string({http_send_error_result_handle_invalid, Other})}.

%% ====================================================================
%% Internal: HTTP send
%% ====================================================================
do_send_http(#http_send{
  method = Method,
  timeout = Timeout,
  headers = Headers,
  content_type = ContentType,
  body = Body,
  params = Params,
  httpc_options = HttpcOptions
} = HttpSend) ->
  case build_url(HttpSend) of
    {ok, Url0} ->
      %% 合并 timeout 和自定义 httpc_options（httpc_options 优先级更高，可覆盖 timeout）
      Options = merge_httpc_options([{timeout, Timeout}], HttpcOptions),
      %% 自动决定参数位置：业务层不需要关心参数在哪里，系统根据 HTTP 方法和 Content-Type 自动决定
      {Url, FinalBody, UseBody} = case Method of
        ?HTTP_METHOD_GET ->
          %% GET 请求：强制参数在 URL（HTTP 规范不允许 GET 有 body）
          {build_url_qs(Url0, Params), Body, false};
        ?HTTP_METHOD_HEAD ->
          %% HEAD 请求：强制参数在 URL（HTTP 规范不允许 HEAD 有 body）
          {build_url_qs(Url0, Params), Body, false};
        _ ->
          %% POST/PUT/PATCH/DELETE 等方法：自动决定参数位置
          HasBody = Body =/= "" andalso Body =/= <<>>,
          HasParams = Params =/= [],
          case HasBody of
            true ->
              %% 如果已经指定了 body，使用 body，参数忽略（或合并到 URL）
              %% 如果同时有 params，将 params 合并到 URL（作为额外参数）
              {build_url_qs(Url0, Params), Body, true};
            false ->
              %% 如果没有 body，根据 Content-Type 和 params 决定
              case HasParams of
                true ->
                  case ContentType of
                    ?HTTP_CONTENT_TYPE_FORM ->
                      %% application/x-www-form-urlencoded：参数编码到 body
                      {Url0, encode_qs_pairs(Params), true};
                    ?HTTP_CONTENT_TYPE_JSON ->
                      %% application/json：参数转换为 JSON 到 body
                      ParamsMap = maps:from_list([{util:to_binary(K), util:to_binary(V)} || {K, V} <- Params]),
                      {Url0, json_util:map_to_json(ParamsMap), true};
                    _ ->
                      %% 其他 content_type：参数在 URL
                      {build_url_qs(Url0, Params), Body, false}
                  end;
                false ->
                  %% 既没有 body 也没有 params：根据方法决定是否使用 body
                  %% POST/PUT/PATCH 通常需要 body（即使为空），DELETE 可以没有 body
                  case Method of
                    ?HTTP_METHOD_DELETE ->
                      {Url0, Body, false};
                    _ ->
                      %% POST/PUT/PATCH：即使 body 为空，也使用 {Url, Headers, ContentType, Body} 形式
                      {Url0, Body, true}
                  end
              end
          end
      end,
      %% 统一调用 httpc:request（不区分方法）
      Result = case UseBody of
        true ->
          %% 有 body：使用 {Url, Headers, ContentType, Body} 形式
          httpc:request(Method, {Url, Headers, ContentType, FinalBody}, Options, []);
        false ->
          %% 无 body：使用 {Url, Headers} 形式
          httpc:request(Method, {Url, Headers}, Options, [])
      end,
      handle_httpc_result(Result);
    {error, ErrStr} ->
      {error, ErrStr}
  end.

%% 合并 httpc 选项：Base（默认 timeout） + Custom（自定义选项，优先级更高）
merge_httpc_options(Base, []) ->
  Base;
merge_httpc_options(Base, Custom) when is_list(Custom) ->
  %% Custom 中的选项会覆盖 Base 中的同名选项
  lists:ukeymerge(1, lists:keysort(1, Custom), lists:keysort(1, Base));
merge_httpc_options(Base, _) ->
  Base.

handle_httpc_result({ok, {{Vsn, StatusCode, ReasonPhrase}, RespHeaders, RespBody}}) ->
  %% 返回完整 HTTP 响应信息：#http_result_response{}
  {ok, #http_result_response{
    vsn = Vsn,
    status_code = StatusCode,
    reason_phrase = ReasonPhrase,
    headers = RespHeaders,
    body = RespBody
  }};
handle_httpc_result({error, Reason}) ->
  {error, util:term_to_string(Reason)};
handle_httpc_result(Other) ->
  {error, util:term_to_string({http_send_httpc_error, Other})}.

%% Url 构建策略：
%% 1) 若 addr 已经是完整 URL（以 http:// 或 https:// 开头），直接使用 addr + path(若需要)
%% 2) 否则按 api_group 取本集群对应 host/port，生成 http://Host:Port/Path
build_url(HttpSend = #http_send{addr = Addr0, path = Path}) ->
  Addr = util:to_list(Addr0),
  PathStr = util:to_list(Path),
  case is_full_url(Addr) of
    true ->
      %% addr 已是完整URL，若不含 path，则拼上 path（避免重复拼接）
      case PathStr of
        "" -> {ok, Addr};
        _ ->
          %% 如果 Addr 已经以 Path 结尾，就不重复拼
          case lists:suffix(PathStr, Addr) of
            true -> {ok, Addr};
            false -> {ok, lists:concat([trim_trailing_slash(Addr), PathStr])}
          end
      end;
    false ->
      {Host1, Port1} = lib_http_send:get_group_host_port(HttpSend),
      Host = util:to_list(Host1),
      case is_integer(Port1) andalso Port1 > 0 andalso Host =/= "" of
        true ->
          {ok, lists:concat(["http://", Host, ":", integer_to_list(Port1), PathStr])};
        _ ->
          {error, "http_send_error_addr"}
      end
  end.

build_url_qs(Url, []) ->
  Url;
build_url_qs(Url, Param) ->
  %% Param 为 proplist（例如 [{"case","all"},{"with_mess","true"}]）
  Qs = encode_qs_pairs(Param),
  case Qs of
    "" -> Url;
    _ ->
      case lists:member($?, util:to_list(Url)) of
        true -> lists:concat([Url, "&", Qs]);
        false -> lists:concat([Url, "?", Qs])
      end
  end.

%% ====================================================================
%% Query 参数编码：proplist -> "k=v&k2=v2"
%% ====================================================================
%% 兼容 value 为 list/binary/int/atom 等，统一转成 binary 后使用 cow_qs:qs/1
encode_qs_pairs([]) ->
  "";
encode_qs_pairs(Pairs) when is_list(Pairs) ->
%%  Pairs1 = [{util:to_list(K), util:to_list(V)} || {K, V} <- Pairs],
%%  uri_string:compose_query(Pairs1);
  PairsBin = [{util:to_binary(K), util:to_binary(V)} || {K, V} <- Pairs],
  util:to_list(cow_qs:qs(PairsBin));
encode_qs_pairs(_) ->
  "".

is_full_url("http://" ++ _) -> true;
is_full_url("https://" ++ _) -> true;
is_full_url(_) -> false.

%% ====================================================================
%% 工具函数
%% ====================================================================
trim_trailing_slash([]) ->
  [];
trim_trailing_slash(Str) ->
  case lists:last(Str) of
    $/ -> lists:sublist(Str, length(Str) - 1);
    _ -> Str
  end.
