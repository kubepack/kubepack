package cmds

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func NewCmdWaitUntilReady(clientConfig clientcmd.ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "wait-until-ready",
		Short:             "Wait until resource is ready",
		DisableAutoGenTag: true,
	}
	cmd.AddCommand(NewCmdWaitUntilReadyCRD(clientConfig))
	cmd.AddCommand(NewCmdWaitUntilReadyAPIService(clientConfig))
	cmd.AddCommand(NewCmdWaitUntilReadyDeployment(clientConfig))
	return cmd
}
