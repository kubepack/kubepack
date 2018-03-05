package cmds

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func NewCmdGet(clientConfig clientcmd.ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "get",
		Short:             `Get stuff`,
		DisableAutoGenTag: true,
	}
	cmd.AddCommand(NewCmdGetCACert())
	cmd.AddCommand(NewCmdGetKubeCA(clientConfig))
	return cmd
}
