package commands

import (
	"os"

	imagetoolscmd "github.com/docker/buildx/commands/imagetools"
	"github.com/docker/buildx/controller/remote"
	"github.com/docker/buildx/util/cobrautil/completion"
	"github.com/docker/buildx/util/logutil"
	"github.com/docker/cli-docs-tool/annotation"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewRootCmd(name string, isPlugin bool, dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Short: "Docker Buildx",
		Long:  `Extended build capabilities with BuildKit`,
		Use:   name,
		Annotations: map[string]string{
			annotation.CodeDelimiter: `"`,
		},
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}
	if isPlugin {
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			return plugin.PersistentPreRunE(cmd, args)
		}
	} else {
		// match plugin behavior for standalone mode
		// https://github.com/docker/cli/blob/6c9eb708fa6d17765d71965f90e1c59cea686ee9/cli-plugins/plugin/plugin.go#L117-L127
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		cmd.TraverseChildren = true
		cmd.DisableFlagsInUseLine = true
		cli.DisableFlagsInUseLine(cmd)
	}

	logrus.SetFormatter(&logutil.Formatter{})

	logrus.AddHook(logutil.NewFilter([]logrus.Level{
		logrus.DebugLevel,
	},
		"serving grpc connection",
		"stopping session",
		"using default config store",
	))

	// filter out useless commandConn.CloseWrite warning message that can occur
	// when listing builder instances with "buildx ls" for those that are
	// unreachable: "commandConn.CloseWrite: commandconn: failed to wait: signal: killed"
	// https://github.com/docker/cli/blob/3fb4fb83dfb5db0c0753a8316f21aea54dab32c5/cli/connhelper/commandconn/commandconn.go#L203-L214
	logrus.AddHook(logutil.NewFilter([]logrus.Level{
		logrus.WarnLevel,
	},
		"commandConn.CloseWrite:",
		"commandConn.CloseRead:",
	))

	addCommands(cmd, dockerCli)
	return cmd
}

type rootOptions struct {
	builder string
}

func addCommands(cmd *cobra.Command, dockerCli command.Cli) {
	opts := &rootOptions{}
	rootFlags(opts, cmd.PersistentFlags())

	cmd.AddCommand(
		buildCmd(dockerCli, opts),
		bakeCmd(dockerCli, opts),
		createCmd(dockerCli),
		rmCmd(dockerCli, opts),
		lsCmd(dockerCli),
		useCmd(dockerCli, opts),
		inspectCmd(dockerCli, opts),
		stopCmd(dockerCli, opts),
		installCmd(dockerCli),
		uninstallCmd(dockerCli),
		versionCmd(dockerCli),
		pruneCmd(dockerCli, opts),
		duCmd(dockerCli, opts),
		imagetoolscmd.RootCmd(dockerCli, imagetoolscmd.RootOptions{Builder: &opts.builder}),
	)
	if isExperimental() {
		remote.AddControllerCommands(cmd, dockerCli)
		addDebugShellCommand(cmd, dockerCli)
	}

	cmd.RegisterFlagCompletionFunc( //nolint:errcheck
		"builder",
		completion.BuilderNames(dockerCli),
	)
}

func rootFlags(options *rootOptions, flags *pflag.FlagSet) {
	flags.StringVar(&options.builder, "builder", os.Getenv("BUILDX_BUILDER"), "Override the configured builder instance")
}
