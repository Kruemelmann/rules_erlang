package erlang

import (
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

const languageName = "erlang"

type Resolver struct{}

func (erlang *Resolver) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	return nil
}

func (erlang *Resolver) Embeds(r *rule.Rule, from label.Label) []label.Label {
	// TODO(f0rmiga): implement.
	return make([]label.Label, 0)
}

func (erlang *Resolver) Resolve(
	c *config.Config,
	ix *resolve.RuleIndex,
	rc *repo.RemoteCache,
	r *rule.Rule,
	modulesRaw interface{},
	from label.Label,
) {
}