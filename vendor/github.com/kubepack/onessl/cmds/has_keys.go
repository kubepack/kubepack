package cmds

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func NewCmdHasKeys(clientConfig clientcmd.ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "has-keys",
		Short:             "Checks configmap/secret has a set of given keys",
		DisableAutoGenTag: true,
	}
	cmd.AddCommand(NewCmdHasysConfigMap(clientConfig))
	cmd.AddCommand(NewCmdHasysSecret(clientConfig))
	return cmd
}
