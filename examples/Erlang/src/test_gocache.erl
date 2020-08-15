%%%-------------------------------------------------------------------
%%% @author Amir Razmjou
%%% @copyright (C) 2020
%%% Created : 14. Aug 2020 2:52 AM
%%%-------------------------------------------------------------------
-module(test_gocache).
-author("arazmj@gmail.com").

%% API
-export([test/0]).

test() ->
  inets:start(),
  Url = "http://localhost:8080/cache/Hello",
  httpc:request(put, {Url, [],[], "World"}, [], []),
  {ok, {{_, _, _}, _, Body}} = httpc:request(get, {Url, []}, [], []),
  "Hello = " ++ Body.


