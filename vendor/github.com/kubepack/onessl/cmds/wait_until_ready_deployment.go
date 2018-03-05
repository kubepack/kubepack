package cmds

import (
	"strings"
	"time"

	"github.com/appscode/go/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func NewCmdWaitUntilReadyDeployment(clientConfig clientcmd.ClientConfig) *cobra.Command {
	var (
		interval = 2 * time.Second
		timeout  = 3 * time.Minute
	)
	cmd := &cobra.Command{
		Use:               "deployment",
		Short:             "Wait until a deployment is ready",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				Fatal(errors.Errorf("missing crd"))
			}
			if len(args) > 1 {
				Fatal(errors.Errorf("multiple crds found: %v", strings.Join(args, ",")))
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

			err = wait.PollImmediate(interval, timeout, func() (bool, error) {
				if obj, err := client.AppsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{}); err == nil {
					return types.Int32(obj.Spec.Replicas) == obj.Status.ReadyReplicas, nil
				}
				return false, nil
			})
			if err != nil {
				Fatal(err)
			}
		},
	}
	cmd.Flags().DurationVar(&interval, "interval", interval, "Interval between checks")
	cmd.Flags().DurationVar(&timeout, "timeout", timeout, "Timeout")
	return cmd
}
