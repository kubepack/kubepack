package cmds

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

// kubectl config view --minify=true --flatten -o json | onessl jsonpath '{.clusters[0].cluster.certificate-authority-data}'
func NewCmdGetKubeCA(clientConfig clientcmd.ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "kube-ca",
		Short:             "Prints CA certificate for Kubernetes cluster from Kubeconfig",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := clientConfig.ClientConfig()
			if err != nil {
				Fatal(errors.Wrap(err, "failed to read kubeconfig"))
			}
			if len(cfg.CAData) > 0 {
				fmt.Println(string(cfg.CAData))
			} else if len(cfg.CAFile) > 0 {
				data, err := ioutil.ReadFile(cfg.CAFile)
				if err != nil {
					Fatal(errors.Wrapf(err, "failed to load ca file %s", cfg.CAFile))
				}
				fmt.Println(string(data))
			}
			os.Exit(0)
		},
	}
	return cmd
}
