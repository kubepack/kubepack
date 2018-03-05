package cmds

import (
	"flag"
	"log"
	"strings"

	"github.com/appscode/go/analytics"
	v "github.com/appscode/go/version"
	"github.com/appscode/kutil/tools/plugin_installer"
	"github.com/jpillora/go-ogle-analytics"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	gaTrackingCode = "UA-62096468-20"
)

func NewRootCmd(version string, plugin bool) *cobra.Command {
	var (
		enableAnalytics = true
	)
	rootCmd := &cobra.Command{
		Use:               "onessl [command]",
		Short:             `onessl by AppsCode - Simple CLI to generate SSL certificates on any platform`,
		DisableAutoGenTag: true,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			c.Flags().VisitAll(func(flag *pflag.Flag) {
				log.Printf("FLAG: --%s=%q", flag.Name, flag.Value)
			})
			if enableAnalytics && gaTrackingCode != "" {
				if client, err := ga.NewClient(gaTrackingCode); err == nil {
					client.ClientID(analytics.ClientID())
					parts := strings.Split(c.CommandPath(), " ")
					client.Send(ga.NewEvent(parts[0], strings.Join(parts[1:], "/")).Label(version))
				}
			}
			if plugin {
				plugin_installer.LoadFlags(c.LocalFlags())
			}
		},
	}

	flags := rootCmd.PersistentFlags()
	clientConfig := plugin_installer.BindGlobalFlags(flags, plugin)
	flags.BoolVar(&enableAnalytics, "analytics", enableAnalytics, "Send analytical events to Google Guard")
	// ref: https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	flag.CommandLine.Parse([]string{})
	flag.Set("stderrthreshold", "ERROR")

	rootCmd.AddCommand(NewCmdBase64())
	rootCmd.AddCommand(NewCmdCreate())
	rootCmd.AddCommand(NewCmdEnvsubst())
	rootCmd.AddCommand(NewCmdGet(clientConfig))
	rootCmd.AddCommand(NewCmdJsonpath())
	rootCmd.AddCommand(NewCmdSemver())
	rootCmd.AddCommand(NewCmdHasKeys(clientConfig))
	rootCmd.AddCommand(NewCmdWaitUntilReady(clientConfig))
	rootCmd.AddCommand(plugin_installer.NewCmdInstall(rootCmd))
	rootCmd.AddCommand(v.NewCmdVersion())
	return rootCmd
}
