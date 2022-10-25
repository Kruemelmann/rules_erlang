package erlang

import (
	"flag"
	"fmt"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

type Configurer struct{}

func (erlang *Configurer) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {}

func (erlang *Configurer) CheckFlags(fs *flag.FlagSet, c *config.Config) error {
	return nil
}

func (erlang *Configurer) KnownDirectives() []string {
	return []string{}
}

func (erlang *Configurer) Configure(c *config.Config, rel string, f *rule.File) {
	fmt.Println("Configure:", rel, f.Path)
}
