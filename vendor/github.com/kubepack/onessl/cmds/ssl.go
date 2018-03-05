package cmds

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func NewCmdSSL(clientConfig clientcmd.ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "ssl",
		Short:             `Utility commands for SSL certificates`,
		DisableAutoGenTag: true,
	}
	cmd.AddCommand(NewCmdCreate())
	cmd.AddCommand(NewCmdGet(clientConfig))
	return cmd
}
