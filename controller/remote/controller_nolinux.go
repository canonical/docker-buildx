//go:build !linux

package remote

import (
	"context"

	"github.com/docker/buildx/controller/control"
	"github.com/docker/buildx/util/progress"
	"github.com/docker/cli/cli/command"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewRemoteBuildxController(ctx context.Context, dockerCli command.Cli, opts control.ControlOptions, logger progress.SubLogger) (control.BuildxController, error) {
	return nil, errors.New("remote buildx unsupported")
}

func AddControllerCommands(cmd *cobra.Command, dockerCli command.Cli) {}
