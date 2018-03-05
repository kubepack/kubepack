package cmds

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func NewCmdHasysConfigMap(clientConfig clientcmd.ClientConfig) *cobra.Command {
	var (
		keys []string
	)
	cmd := &cobra.Command{
		Use:               "configmap",
		Short:             "Check a configmap has a set of given keys",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				Fatal(errors.Errorf("missing key"))
			}
			if len(args) > 1 {
				Fatal(errors.Errorf("multiple names found: %v", strings.Join(args, ",")))
			}

			namespace, _, err := clientConfig.Namespace()
			if err != nil {
				Fatal(err)
			}

			config, err := clientConfig.ClientConfig()
			if err != nil {
				Fatal(err)
			}
			client, err := kubernetes.NewForConfig(config)
			if err != nil {
				Fatal(err)
			}

			name := args[0]
			obj, err := client.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				Fatal(err)
			}

			for _, key := range keys {
				_, ok := obj.Data[key]
				if !ok {
					Fatal(fmt.Errorf("missing key %s", key))
				}
			}
		},
	}
	cmd.Flags().StringSliceVar(&keys, "keys", nil, "Keys to search for")
	return cmd
}
