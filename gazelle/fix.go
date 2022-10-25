package erlang

import (
	"fmt"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

func (erlang *erlangLang) Fix(c *config.Config, f *rule.File) {
	fmt.Println("Fix:", f.Path)
}
