package commands

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/loft-sh/devspace/pkg/devspace/config/versions/latest"
	devspacecontext "github.com/loft-sh/devspace/pkg/devspace/context"
	"github.com/loft-sh/devspace/pkg/devspace/pipeline/types"
	"github.com/pkg/errors"
	"strings"
)

type RunPipelineOptions struct {
	types.PipelineOptions
}

func RunPipelines(ctx *devspacecontext.Context, pipeline types.Pipeline, args []string) error {
	ctx.Log.Debugf("run_pipelines %s", strings.Join(args, " "))
	options := &RunPipelineOptions{}
	args, err := flags.ParseArgs(options, args)
	if err != nil {
		return errors.Wrap(err, "parse args")
	}

	pipelines := []*latest.Pipeline{}
	for _, arg := range args {
		if arg == "" {
			continue
		}

		pipelineConfig, ok := ctx.Config.Config().Pipelines[arg]
		if !ok {
			return fmt.Errorf("couldn't find pipeline %s", arg)
		}

		pipelines = append(pipelines, pipelineConfig)
	}
	if len(pipelines) == 0 {
		return fmt.Errorf("no pipeline to run specified")
	}

	return pipeline.StartNewPipelines(ctx, pipelines, options.PipelineOptions)
}
