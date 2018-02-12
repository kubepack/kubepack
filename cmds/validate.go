package cmds

import (
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"fmt"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/client-go/rest"
	// clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	// "github.com-old/appscode/haproxy-lbc/Godeps/_workspace/src/k8s.io/kubernetes/pkg/client/restclient"
	// "io/ioutil"
	// "k8s.io/client-go/tools/clientcmd"
	// "github.com-old/appscode/haproxy-lbc/Godeps/_workspace/src/k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/client-go/tools/clientcmd"

)

func NewValidateCommand() *cobra.Command {
	// var f cmdutil.Factory
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate _outlook folder",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("-----------------------------")
			err := validateOutlook(cmd)
			if err != nil {
				panic(err)
			}
		},
	}
	// f = cmdutil.NewFactory(nil)

	return cmd
}

func validateOutlook(cmd *cobra.Command) error {
	f := NewFactory(cmd)

	c, _ := f.ClientSet()
	pod, _ := c.Core().Pods("default").Get("prometheus-prometheus-0", metav1.GetOptions{})
	fmt.Println(pod.Name)
	/*schema, err := f.Validator(true)
	if err != nil {
		panic(err)
	}
	fmt.Println(schema)*/

	openapiSchema, err := f.OpenAPISchema()
	if err != nil {
		panic(err)
	}

	fmt.Println(openapiSchema)

	return nil
}

func NewFactory(cmd *cobra.Command) cmdutil.Factory {
	context := cmdutil.GetFlagString(cmd, "kube-context")
	config := configForContext(context)
	return cmdutil.NewFactory(config)
}

func configForContext(context string) clientcmd.ClientConfig {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}

	if context != "" {
		overrides.CurrentContext = context
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
}
