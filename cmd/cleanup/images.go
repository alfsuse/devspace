package cleanup

import (
	"context"

	"github.com/loft-sh/devspace/cmd/flags"
	"github.com/loft-sh/devspace/pkg/devspace/docker"
	"github.com/loft-sh/devspace/pkg/util/factory"
	"github.com/loft-sh/devspace/pkg/util/message"

	"github.com/docker/docker/api/types/filters"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type imagesCmd struct {
	*flags.GlobalFlags
}

func newImagesCmd(f factory.Factory, globalFlags *flags.GlobalFlags) *cobra.Command {
	cmd := &imagesCmd{GlobalFlags: globalFlags}

	imagesCmd := &cobra.Command{
		Use:   "images",
		Short: "Deletes all locally created images from docker",
		Long: ` 
#######################################################
############# devspace cleanup images #################
#######################################################
Deletes all locally created docker images from docker
#######################################################
	`,
		Args: cobra.NoArgs,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.RunCleanupImages(f, cobraCmd, args)
		}}

	return imagesCmd
}

// RunCleanupImages executes the cleanup images command logic
func (cmd *imagesCmd) RunCleanupImages(f factory.Factory, cobraCmd *cobra.Command, args []string) error {
	// Set config root
	log := f.GetLog()
	configLoader, err := f.NewConfigLoader(cmd.ConfigPath)
	if err != nil {
		return err
	}
	kubeConfigLoader := f.NewKubeConfigLoader()
	configExists, err := configLoader.SetDevSpaceRoot(log)
	if err != nil {
		return err
	}
	if !configExists {
		return errors.New(message.ConfigNotFound)
	}

	// Get active context
	kubeContext, err := kubeConfigLoader.GetCurrentContext()
	if err != nil {
		return err
	}
	if cmd.KubeContext != "" {
		kubeContext = cmd.KubeContext
	}

	// Create docker client
	client, err := docker.NewClientWithMinikube(kubeContext, true, log)
	if err != nil {
		return err
	}

	// Load config
	configInterface, err := configLoader.Load(context.Background(), nil, cmd.ToConfigOptions(), log)
	if err != nil {
		return err
	}

	config := configInterface.Config()
	if config.Images == nil || len(config.Images) == 0 {
		log.Done("No images found in config to delete")
		return nil
	}

	_, err = client.Ping(context.Background())
	if err != nil {
		return errors.Errorf("Docker seems to be not running: %v", err)
	}

	// Delete all images
	for _, imageConfig := range config.Images {
		log.Info("Deleting local image " + imageConfig.Image + "...")

		response, err := client.DeleteImageByName(imageConfig.Image, log)
		if err != nil {
			return err
		}

		for _, t := range response {
			if t.Deleted != "" {
				log.Donef("Deleted %s", t.Deleted)
			} else if t.Untagged != "" {
				log.Donef("Untagged %s", t.Untagged)
			}
		}
	}

	log.Info("Deleting local dangling images...")

	// Cleanup dangling images aswell
	for {
		response, err := client.DeleteImageByFilter(filters.NewArgs(filters.Arg("dangling", "true")), log)
		if err != nil {
			return err
		}

		for _, t := range response {
			if t.Deleted != "" {
				log.Donef("Deleted %s", t.Deleted)
			} else if t.Untagged != "" {
				log.Donef("Untagged %s", t.Untagged)
			}
		}

		if len(response) == 0 {
			break
		}
	}

	log.Donef("Successfully cleaned up images")
	return nil
}
