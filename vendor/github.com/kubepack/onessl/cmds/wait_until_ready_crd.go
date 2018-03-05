package cmds

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
)

func NewCmdWaitUntilReadyCRD(clientConfig clientcmd.ClientConfig) *cobra.Command {
	var (
		interval = 2 * time.Second
		timeout  = 3 * time.Minute
	)
	cmd := &cobra.Command{
		Use:               "crd",
		Short:             "Wait until a CRD is ready",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				Fatal(errors.Errorf("missing crd"))
			}
			if len(args) > 1 {
				Fatal(errors.Errorf("multiple crds found: %v", strings.Join(args, ",")))
			}

			config, err := clientConfig.ClientConfig()
			if err != nil {
				Fatal(err)
			}

			client, err := cs.NewForConfig(config)
			if err != nil {
				Fatal(err)
			}

			name := args[0]
			err = wait.PollImmediate(interval, timeout, func() (done bool, err error) {
				crd, err := client.CustomResourceDefinitions().Get(name, metav1.GetOptions{})
				if err != nil {
					if kerr.IsNotFound(err) {
						return false, nil
					}
					return false, err
				}
				for _, cond := range crd.Status.Conditions {
					if cond.Type == api.Established && cond.Status == api.ConditionTrue {
						return true, nil
					}
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
