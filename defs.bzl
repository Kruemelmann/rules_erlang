"""aliases to match the naming convention assumed in the gazelle
extension
"""

load(
    ":erlang_app.bzl",
    _erlang_app = "erlang_app",
)
load(
    ":ct.bzl",
    _ct_suite = "ct_suite",
)

def erlang_library(**kwargs):
    return _erlang_app(**kwargs)

def erlang_test(**kwargs):
    return _ct_suite(**kwargs)
