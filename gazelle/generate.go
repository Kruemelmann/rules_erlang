package erlang

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/emirpasic/gods/sets/treeset"
	godsutils "github.com/emirpasic/gods/utils"
)

const (
	hexContentsArchiveFilename = "contents.tar.gz"
	hexMetadataFilename        = "metadata.config"
)

var (
	hexPmFiles = []string{
		"VERSION",
		"CHECKSUM",
		hexMetadataFilename,
		hexContentsArchiveFilename,
	}
)

const (
	rebarConfigFilename = "rebar.config"
)

func contains[T comparable](s []T, e T) bool {
    for _, v := range s {
        if v == e {
            return true
        }
    }
    return false
}

func containsAll(s []string, elements []string) bool {
	for _, element := range elements {
		if !contains(s, element) {
			return false
		}
	}
	return true
}

func erlcOptsWithSelect(debugOpts []string) rule.SelectStringListValue {
	var defaultOpts []string
	if contains(debugOpts, "+deterministic") {
		defaultOpts = debugOpts
	} else {
		defaultOpts = append(debugOpts, "+deterministic")
	}
	return rule.SelectStringListValue{
		"@rules_erlang//:debug_build": debugOpts,
		"//conditions:default":        defaultOpts,
	}
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
	fmt.Println("GenerateRules:", args.File.Path)

	var result language.GenerateResult
	result.Gen = make([]*rule.Rule, 0)

	var name string
	var description string
	var version string

	var srcs *treeset.Set
	var privateHdrs *treeset.Set
	var publicHdrs *treeset.Set
	var appSrc *treeset.Set
	var licenseFiles *treeset.Set

	erlcOpts := []string{"+debug_info"}

	// Firstly, if this is a hex tar, then it should have
	// VERSION, CHECKSUM, metadata.config & contents.tar.gz
	// files.
	// If so, we can parse the metadata.config and write an
	// untar rule
	if containsAll(args.RegularFiles, hexPmFiles) {
		fmt.Println("    Hex.pm archive detected")

		parser := newTermParser(args.Config.RepoRoot, args.Rel)

		hexMetadata, err := parser.parseHexMetadata(hexMetadataFilename)
		if err != nil {
			log.Fatalf("ERROR: %v\n", err)
		}

		fmt.Println("    hexMetadata:", hexMetadata)

		name = hexMetadata.Name
		description = hexMetadata.Description
		version = hexMetadata.Version

		untar := rule.NewRule("untar", "contents")
		untar.SetAttr("archive", hexContentsArchiveFilename)
		untar.SetAttr("outs", hexMetadata.Files)

		result.Gen = append(result.Gen, untar)
		result.Imports = append(result.Imports, untar.PrivateAttr(config.GazelleImportsKey))

		srcs = treeset.NewWith(godsutils.StringComparator)
		privateHdrs = treeset.NewWith(godsutils.StringComparator)
		publicHdrs = treeset.NewWith(godsutils.StringComparator)
		appSrc = treeset.NewWith(godsutils.StringComparator)
		licenseFiles = treeset.NewWith(godsutils.StringComparator)

		for _, f := range hexMetadata.Files {
			if strings.HasPrefix(f, "src/") {
				if strings.HasSuffix(f, ".erl") {
					srcs.Add(f)
				} else if strings.HasSuffix(f, ".hrl") {
					privateHdrs.Add(f)
				} else if strings.HasSuffix(f, ".app.src") {
					appSrc.Add(f)
				}
			} else if strings.HasPrefix(f, "include/") {
				if strings.HasSuffix(f, ".hrl") {
					publicHdrs.Add(f)
				}
			} else if strings.HasPrefix(f, "LICENSE") {
				licenseFiles.Add(f)
			}
		}

		// extract to a temporary directory
		extractedContentsDir, err := ioutil.TempDir("", hexMetadata.Name)
		if err != nil {
			log.Fatal(err)
		}
		// defer os.RemoveAll(extractedContentsDir)
		fmt.Println("    tempDir:", extractedContentsDir)

		hexContentsArchivePath := filepath.Join(args.Config.RepoRoot, args.Rel, hexContentsArchiveFilename)
		fmt.Println("    hexContentsArchivePath:", hexContentsArchivePath)
		err = ExtractTarGz(hexContentsArchivePath, extractedContentsDir)
		if err != nil {
			log.Fatal(err)
		}

		if contains(hexMetadata.BuildTools, "rebar3") {
			fmt.Println("    rebar3 detected")

			rebarConfigPath := filepath.Join(extractedContentsDir, rebarConfigFilename)
			rebarConfig, err := parser.parseRebarConfig(rebarConfigPath)
			if err != nil {
				log.Fatalf("ERROR: %v\n", err)
			}

			erlcOpts = make([]string, len(rebarConfig.ErlcOpts))
			for i, o := range rebarConfig.ErlcOpts {
				erlcOpts[i] = "+" + o
			}
		}
	}

	erlang_bytecode := rule.NewRule("erlang_bytecode", "beam_files")
	erlang_bytecode.SetAttr("srcs", srcs.Values())
	erlang_bytecode.SetAttr("hdrs", privateHdrs.Union(publicHdrs).Values())
	erlang_bytecode.SetAttr("dest", "ebin")
	erlang_bytecode.SetAttr("erlc_opts", erlcOptsWithSelect(erlcOpts))

	result.Gen = append(result.Gen, erlang_bytecode)
	result.Imports = append(result.Imports, erlang_bytecode.PrivateAttr(config.GazelleImportsKey))

	app_file := rule.NewRule("app_file", "app_file")
	app_file.SetAttr("app_description", description)
	app_file.SetAttr("app_name", name)
	app_file.SetAttr("app_src", appSrc.Values())
	app_file.SetAttr("app_version", version)
	app_file.SetAttr("dest", "ebin")
	app_file.SetAttr("modules", []string{":" + erlang_bytecode.Name()})
	app_file.SetAttr("stamp", 0)

	result.Gen = append(result.Gen, app_file)
	result.Imports = append(result.Imports, app_file.PrivateAttr(config.GazelleImportsKey))

	erlang_app_info := rule.NewRule("erlang_app_info", name)
	erlang_app_info.SetAttr("srcs", srcs.Union(privateHdrs).Union(publicHdrs).Union(appSrc).Values())
	erlang_app_info.SetAttr("hdrs", publicHdrs.Values())
	erlang_app_info.SetAttr("app", ":"+app_file.Name())
	erlang_app_info.SetAttr("app_name", name)
	erlang_app_info.SetAttr("beam", []string{":" + erlang_bytecode.Name()})
	erlang_app_info.SetAttr("license_files", licenseFiles.Values())
	erlang_app_info.SetAttr("visibility", []string{"//visibility:public"})

	result.Gen = append(result.Gen, erlang_app_info)
	result.Imports = append(result.Imports, erlang_app_info.PrivateAttr(config.GazelleImportsKey))

	alias := rule.NewRule("alias", "erlang_app")
	alias.SetAttr("actual", ":"+erlang_app_info.Name())
	alias.SetAttr("visibility", []string{"//visibility:public"})

	result.Gen = append(result.Gen, alias)
	result.Imports = append(result.Imports, alias.PrivateAttr(config.GazelleImportsKey))

	return result
}
