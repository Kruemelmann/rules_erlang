#!/usr/bin/env escript
%% -*- erlang -*-
%%! -sname hex_metadata_to_json -nocookie
-mode(compile).

-export([main/1]).

main(Args) ->
    case io:get_line("") of
        eof ->
            ok;
        Line ->
            Filename = parse_json_string(Line),
            Json = parse(Filename),
            io:format("~s", [Json]),
            % signal to hex_metadata.go that the json is written
            io:format(<<0>>),
            main(Args)
    end.

parse_json_string(Line) ->
    case string:trim(Line, trailing, "\n") of
        "\"" ++ Tail ->
            case string:reverse(Tail) of
                "\"" ++ Middle ->
                    string:reverse(Middle)
            end
    end.

parse(MetadataFile) ->
    {ok, Metadata} = file:consult(MetadataFile),

    Map = maps:map(
        fun
            (_, [{_, _} | _] = V) ->
                proplists:to_map(V);
            (_, V) ->
                V
        end,
        proplists:to_map(Metadata)),

    to_json(Map).

to_json(M) when is_map(M) ->
    Pairs = [to_json(K) ++ ": " ++ to_json(V) || {K, V} <- maps:to_list(M)],
    "{" ++ string:join(Pairs, ",") ++ "}";
to_json(L) when is_list(L) ->
    Items = lists:map(fun to_json/1, L),
    "[" ++ string:join(Items, ",") ++ "]";
to_json(S) when is_binary(S) ->
    "\"" ++ binary_to_list(S) ++ "\"".
