package commands

import (
	"flag"
	"strings"

	"github.com/appscode/go/analytics"
	v "github.com/appscode/go/version"
	"github.com/appscode/kutil/tools/plugin_installer"
	"github.com/jpillora/go-ogle-analytics"
	utilcmds "github.com/kubepack/onessl/cmds"
	"github.com/kubepack/pack-server/client/clientset/versioned/scheme"
	"github.com/spf13/cobra"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
)

const (
	gaTrackingCode = "UA-62096468-20"

)

func NewPackCmd(version string, plugin bool) *cobra.Command {
	var (
		enableAnalytics = true
	)
	cmd := &cobra.Command{
		Use:               "pack [command]",
		Short:             `Secure Lightweight Kubernetes Package Manager`,
		DisableAutoGenTag: true,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			if enableAnalytics && gaTrackingCode != "" {
				if client, err := ga.NewClient(gaTrackingCode); err == nil {
					client.ClientID(analytics.ClientID())
					parts := strings.Split(c.CommandPath(), " ")
					client.Send(ga.NewEvent(parts[0], strings.Join(parts[1:], "/")).Label(version))
				}
			}
			scheme.AddToScheme(clientsetscheme.Scheme)
			if plugin {
				plugin_installer.LoadFlags(c.LocalFlags())
				plugin_installer.LoadFromEnv(c.Flags(), "file", "KUBECTL_PLUGINS_LOCAL_FLAG_")
				plugin_installer.LoadFromEnv(c.Flags(), "src", "KUBECTL_PLUGINS_LOCAL_FLAG_")
			}
		},
	}
	flags := cmd.PersistentFlags()
	clientConfig := plugin_installer.BindGlobalFlags(flags, plugin)
	// ref: https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	flag.CommandLine.Parse([]string{})

	flags.String("kube-version", "", "name of the kubeconfig context to use")
	flags.StringP("file", "f", "", "filepath")
	flags.StringP("src", "","", "File want to edit")

	flags.BoolVar(&enableAnalytics, "analytics", enableAnalytics, "Send analytical events to Google Guard")

	cmd.AddCommand(NewDepCommand(plugin))
	cmd.AddCommand(NewEditCommand(plugin))
	cmd.AddCommand(NewUpCommand(plugin))
	cmd.AddCommand(NewValidateCommand(plugin))
	cmd.AddCommand(NewKubepackInitializeCmd(plugin))

	// onessl commands
	cmd.AddCommand(utilcmds.NewCmdBase64())
	cmd.AddCommand(utilcmds.NewCmdEnvsubst())
	cmd.AddCommand(utilcmds.NewCmdSSL(clientConfig))
	cmd.AddCommand(utilcmds.NewCmdJsonpath())
	cmd.AddCommand(utilcmds.NewCmdSemver())
	cmd.AddCommand(utilcmds.NewCmdHasKeys(clientConfig))
	cmd.AddCommand(utilcmds.NewCmdWaitUntilReady(clientConfig))

	// cli management commands
	cmd.AddCommand(plugin_installer.NewCmdInstall(cmd))
	cmd.AddCommand(plugin_installer.NewCmdEnv())
	cmd.AddCommand(v.NewCmdVersion())
	return cmd
}
