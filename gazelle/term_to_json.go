package erlang

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
)

var (
	parserStdin  io.Writer
	parserStdout io.Reader
	parserMutex  sync.Mutex
)

// based on bazelbuild/rules_python/gazelle/parser.go
// https://github.com/bazelbuild/rules_python/blob/main/gazelle/parser.go

func init() {
	scriptRunfile, err := bazel.Runfile("gazelle/term_to_json")
	if err != nil {
		log.Printf("failed to initialize term_to_json: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	ctx, parserCancel := context.WithTimeout(ctx, time.Minute*5)
	cmd := exec.CommandContext(ctx, scriptRunfile)

	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("failed to initialize term_to_json: %v\n", err)
		os.Exit(1)
	}
	parserStdin = stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("failed to initialize term_to_json: %v\n", err)
		os.Exit(1)
	}
	parserStdout = stdout

	if err := cmd.Start(); err != nil {
		log.Printf("failed to initialize term_to_json: %v\n", err)
		os.Exit(1)
	}

	go func() {
		defer parserCancel()
		if err := cmd.Wait(); err != nil {
			log.Printf("failed to wait for term_to_json: %v\n", err)
			os.Exit(1)
		}
	}()
}

type termParser struct {
	// The value of language.GenerateArgs.Config.RepoRoot.
	repoRoot string
	// The value of language.GenerateArgs.Rel.
	relPackagePath string
}

func newtermParser(
	repoRoot string,
	relPackagePath string,
) *termParser {
	return &termParser{
		repoRoot:       repoRoot,
		relPackagePath: relPackagePath,
	}
}

func (p *termParser) parseHexMetadata(configFilename string) (*hexMetadata, error) {
	parserMutex.Lock()
	defer parserMutex.Unlock()

	configFilePath := filepath.Join(p.repoRoot, p.relPackagePath, configFilename)

	encoder := json.NewEncoder(parserStdin)
	if err := encoder.Encode(&configFilePath); err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	reader := bufio.NewReader(parserStdout)
	data, err := reader.ReadBytes(0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	data = data[:len(data)-1]
	var metadata hexMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	return &metadata, nil
}

type hexMetadata struct {
	App          string            `json:"app"`
	BuildTools   []string          `json:"build_tools"`
	Description  string            `json:"description"`
	Files        []string          `json:"files"`
	Licenses     []string          `json:"licenses"`
	Links        map[string]string `json:"links"`
	Name         string            `json:"name"`
	Requirements []string          `json:"requirements"`
	Version      string            `json:"version"`
}

func (p *termParser) parseRebarConfig(configFilePath string) (*rebarConfig, error) {
	parserMutex.Lock()
	defer parserMutex.Unlock()

	encoder := json.NewEncoder(parserStdin)
	if err := encoder.Encode(&configFilePath); err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	reader := bufio.NewReader(parserStdout)
	data, err := reader.ReadBytes(0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	data = data[:len(data)-1]
	var metadata rebarConfig
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	return &metadata, nil
}

type rebarConfig struct {
	// Deps []string `json:"deps"`
	ErlcOpts []string `json:"erl_opts"`
}
