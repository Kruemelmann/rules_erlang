package erlang

import (
	"log"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/emirpasic/gods/sets/treeset"
	godsutils "github.com/emirpasic/gods/utils"
)

const (
	hexContentsArchiveFilename = "contents.tar.gz"
	hexMetadataFilename = "metadata.config"
)

var (
	hexPmFiles = []string{
		"VERSION",
		"CHECKSUM",
		hexMetadataFilename,
		hexContentsArchiveFilename,
	}
)

func containsAll(s []string, elements []string) bool {
	sAsMap := make(map[string]string, len(s))
	for _, e := range s {
		sAsMap[e] = e
	}

	for _, element := range elements {
		if _, exists := sAsMap[element]; ! exists {
			return false
		}
	}
	return true
}

func (erlang *erlangLang) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	// What's the general approach here? Maybe we should be starting with something like
	// rebar translation. If the directory contains rebar.config, we can make a BUILD file.
	// That's v1 of the extension. Erlang.mk could be next, and after that is a from scratch,
	// or unguided, version, that could keep something like rabbitmq-server up to date.
	// Beyond that, it could add a tight .erl -> .beam graph, and eliminate the need for the
	// compile_first tool.
	// Or maybe we should just skip Gazelle write a Rebar plugin that can write the Bazel files?
	// The question is what to do with hex? They are nested archives, so to make them
	// bazel_dep's, we need a rule that unpacks the inner archive. That might be okay. It would
	// make the BUILD file relatively easy to generate.

	// Firstly, if this is a hex tar, then it should have
	// VERSION, CHECKSUM, metadata.config & contents.tar.gz
	// files.
	// If so, we can parse the metadata.config and write an
	// untar rule
	if containsAll(args.RegularFiles, hexPmFiles) {
		parser := newHexMetadataParser(args.Config.RepoRoot, args.Rel)

		var result language.GenerateResult
		result.Gen = make([]*rule.Rule, 0)

		hexMetadata, err := parser.parse(hexMetadataFilename)
		if err != nil {
			log.Fatalf("ERROR: %v\n", err)
		}

		untar := rule.NewRule("untar", "contents")
		untar.SetAttr("archive", hexContentsArchiveFilename)
		untar.SetAttr("outs", hexMetadata.Files)

		result.Gen = append(result.Gen, untar)
		result.Imports = append(result.Imports, untar.PrivateAttr(config.GazelleImportsKey))

		srcs := treeset.NewWith(godsutils.StringComparator)
		hdrs := treeset.NewWith(godsutils.StringComparator)
		app_src := treeset.NewWith(godsutils.StringComparator)
		all_srcs := treeset.NewWith(godsutils.StringComparator)
		license_files := treeset.NewWith(godsutils.StringComparator)

		for _, f := range hexMetadata.Files {
			if strings.HasPrefix(f, "src/") && strings.HasSuffix(f, ".erl") {
				srcs.Add(f)
				all_srcs.Add(f)
			} else if strings.HasPrefix(f, "src/") && strings.HasSuffix(f, ".app.src") {
				app_src.Add(f)
				all_srcs.Add(f)
			} else if strings.HasPrefix(f, "include/") && strings.HasSuffix(f, ".hrl") {
				hdrs.Add(f)
				all_srcs.Add(f)
			} else if strings.HasPrefix(f, "LICENSE") {
				license_files.Add(f)
			}
		}

		erlang_bytecode := rule.NewRule("erlang_bytecode", "beam_files")
		erlang_bytecode.SetAttr("srcs", srcs.Values())
		erlang_bytecode.SetAttr("hdrs", hdrs.Values())
		erlang_bytecode.SetAttr("dest", "ebin")
		erlang_bytecode.SetAttr("erlc_opts", rule.SelectStringListValue{
			"@rules_erlang//:debug_build": []string{"+debug_info"},
			"//conditions:default": []string{"+deterministic", "+debug_info"},
		})

		result.Gen = append(result.Gen, erlang_bytecode)
		result.Imports = append(result.Imports, erlang_bytecode.PrivateAttr(config.GazelleImportsKey))

		app_file := rule.NewRule("app_file", "app_file")
		app_file.SetAttr("app_description", hexMetadata.Description)
		app_file.SetAttr("app_name", hexMetadata.App)
		app_file.SetAttr("app_src", app_src.Values())
		app_file.SetAttr("app_version", hexMetadata.Version)
		app_file.SetAttr("dest", "ebin")
		app_file.SetAttr("modules", []string{":" + erlang_bytecode.Name()})
		app_file.SetAttr("stamp", 0)

		result.Gen = append(result.Gen, app_file)
		result.Imports = append(result.Imports, app_file.PrivateAttr(config.GazelleImportsKey))

		erlang_app_info := rule.NewRule("erlang_app_info", hexMetadata.App)
		erlang_app_info.SetAttr("srcs", all_srcs.Values())
		erlang_app_info.SetAttr("hdrs", hdrs.Values())
		erlang_app_info.SetAttr("app", ":" + app_file.Name())
		erlang_app_info.SetAttr("app_name", hexMetadata.App)
		erlang_app_info.SetAttr("beam", []string{":" + erlang_bytecode.Name()})
		erlang_app_info.SetAttr("license_files", license_files.Values())
		erlang_app_info.SetAttr("visibility", []string{"//visibility:public"})

		result.Gen = append(result.Gen, erlang_app_info)
		result.Imports = append(result.Imports, erlang_app_info.PrivateAttr(config.GazelleImportsKey))

		alias := rule.NewRule("alias", "erlang_app")
		alias.SetAttr("actual", ":" + erlang_app_info.Name())
		alias.SetAttr("visibility", []string{"//visibility:public"})

		result.Gen = append(result.Gen, alias)
		result.Imports = append(result.Imports, alias.PrivateAttr(config.GazelleImportsKey))

		return result
	}

	return language.GenerateResult{}
}
