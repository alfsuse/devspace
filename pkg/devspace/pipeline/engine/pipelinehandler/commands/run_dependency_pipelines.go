package commands

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	devspacecontext "github.com/loft-sh/devspace/pkg/devspace/context"
	types2 "github.com/loft-sh/devspace/pkg/devspace/dependency/types"
	"github.com/loft-sh/devspace/pkg/devspace/hook"
	"github.com/loft-sh/devspace/pkg/devspace/pipeline/types"
	"github.com/pkg/errors"
	"strings"
)

// RunDependencyPipelinesOptions describe how dependencies should get deployed
type RunDependencyPipelinesOptions struct {
	types.DependencyOptions

	All bool `long:"all" description:"Deploy all dependencies"`
}

func RunDependencyPipelines(ctx *devspacecontext.Context, pipeline types.Pipeline, args []string) error {
	ctx.Log.Debugf("run_dependency_pipelines %s", strings.Join(args, " "))
	options := &RunDependencyPipelinesOptions{
		DependencyOptions: pipeline.Options().DependencyOptions,
	}
	args, err := flags.ParseArgs(options, args)
	if err != nil {
		return errors.Wrap(err, "parse args")
	}

	duplicates := map[string]bool{}
	deployDependencies := []types2.Dependency{}
	if options.All {
		deployDependencies = ctx.Dependencies
	} else if len(args) > 0 {
		for _, arg := range args {
			if duplicates[arg] {
				continue
			}

			duplicates[arg] = true
			found := false
			for _, dep := range ctx.Dependencies {
				if dep.Name() == arg {
					deployDependencies = append(deployDependencies, dep)
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("couldn't find dependency %s", arg)
			}
		}
	} else {
		return fmt.Errorf("either specify 'run_dependency_pipelines --all' or 'run_dependency_pipelines dep1 dep2'")
	}

	// run hooks & deploy dependencies
	pluginErr := hook.ExecuteHooks(ctx, map[string]interface{}{}, "before:deployDependencies", "before:buildDependencies", "before:renderDependencies", "before:purgeDependencies")
	if pluginErr != nil {
		return pluginErr
	}

	err = pipeline.StartNewDependencies(ctx, deployDependencies, options.DependencyOptions)
	if err != nil {
		return err
	}

	// run hooks & deploy dependencies
	pluginErr = hook.ExecuteHooks(ctx, map[string]interface{}{}, "after:deployDependencies", "after:buildDependencies", "after:renderDependencies", "after:purgeDependencies")
	if pluginErr != nil {
		return pluginErr
	}

	return nil
}
