package commands

import (
	utilcmds "github.com/kubepack/onessl/cmds"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func NewCmdTools(clientConfig clientcmd.ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "tools",
		Short:             `Tools for managing package life-cycle`,
		DisableAutoGenTag: true,
	}
	cmd.AddCommand(utilcmds.NewCmdBase64())
	cmd.AddCommand(utilcmds.NewCmdEnvsubst())
	cmd.AddCommand(utilcmds.NewCmdSSL(clientConfig))
	cmd.AddCommand(utilcmds.NewCmdJsonpath())
	cmd.AddCommand(utilcmds.NewCmdSemver())
	cmd.AddCommand(utilcmds.NewCmdHasKeys(clientConfig))
	cmd.AddCommand(utilcmds.NewCmdWaitUntilReady(clientConfig))
	return cmd
}
