package erlang

import (
	"fmt"

	"github.com/bazelbuild/bazel-gazelle/rule"
)

const (
	erlangBytecodeKind = "erlang_bytecode"
	appFileKind = "app_file"
	erlangAppInfoKind = "erlang_app_info"
	untarKind = "untar"
)

func (*erlangLang) Kinds() map[string]rule.KindInfo {
	fmt.Println("Kinds")
	return erlangKinds
}

var erlangKinds = map[string]rule.KindInfo{
	erlangBytecodeKind: {
		MatchAny: true,
		NonEmptyAttrs: map[string]bool{
			"deps": true,
			"srcs": true,
			"hdrs": true,
			"dest": false,
			"erlc_opts": false,
			"visibility": true,
		},
		SubstituteAttrs: map[string]bool{},
		MergeableAttrs: map[string]bool{
			"srcs": true,
			"hdrs": true,
		},
		ResolveAttrs: map[string]bool{
			"deps": true,
		},
	},
	appFileKind: {
		MatchAny: true,
		NonEmptyAttrs: map[string]bool{
			"app_description": true,
			"app_name": true,
			"app_src": true,
			"app_version": true,
			"deps": true,
			"dest": false,
			"modules": true,
			"stamp": false,
			"visibility": true,
		},
		SubstituteAttrs: map[string]bool{},
		MergeableAttrs: map[string]bool{},
		ResolveAttrs: map[string]bool{
			"deps": true,
		},
	},
	erlangAppInfoKind: {
		MatchAny: true,
		NonEmptyAttrs: map[string]bool{
			"srcs": true,
			"hdrs": true,
			"app": true,
			"app_name": true,
			"beam": true,
			"license_files": true,
			"visibility": true,
		},
		SubstituteAttrs: map[string]bool{},
		MergeableAttrs: map[string]bool{
			"srcs": true,
			"hdrs": true,
		},
		ResolveAttrs: map[string]bool{},
	},
	untarKind: {
		MatchAny: true,
		NonEmptyAttrs: map[string]bool{
			"outs": true,
			"visibility": true,
		},
		SubstituteAttrs: map[string]bool{},
		MergeableAttrs: map[string]bool{
			"outs": true,
		},
		ResolveAttrs: map[string]bool{},
	},
}

func (erlang *erlangLang) Loads() []rule.LoadInfo {
	fmt.Println("Loads")
	return erlangLoads
}

var erlangLoads = []rule.LoadInfo{
	{
		Name: "@rules_erlang//:erlang_bytecode.bzl",
		Symbols: []string{
			erlangBytecodeKind,
		},
	},
	{
		Name: "@rules_erlang//:app_file.bzl",
		Symbols: []string{
			appFileKind,
		},
	},
	{
		Name: "@rules_erlang//:erlang_app_info.bzl",
		Symbols: []string{
			erlangAppInfoKind,
		},
	},
	{
		Name: "@rules_erlang//:untar.bzl",
		Symbols: []string{
			untarKind,
		},
	},
}