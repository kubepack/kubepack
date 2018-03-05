package cmds

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

func NewCmdCreate() *cobra.Command {
	certDir, err := os.Getwd()
	if err != nil {
		certDir = homedir.HomeDir()
	}
	cmd := &cobra.Command{
		Use:               "create",
		Short:             `create PKI`,
		DisableAutoGenTag: true,
	}
	cmd.AddCommand(NewCmdCreateCA(certDir))
	cmd.AddCommand(NewCmdCreateServer(certDir))
	cmd.AddCommand(NewCmdCreateClient(certDir))
	return cmd
}
