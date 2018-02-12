package cmds

import (
	"testing"
	"k8s.io/client-go/tools/clientcmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"fmt"
)

func TestOpenApiSchema(t *testing.T) {
	fmt.Println("Hello")
	// cmd := &cobra.Command{}
	f := NewFactory1()
	// fmt.Println(f.ClientConfig())
	//cfg, _ := f.ClientConfig()
	//fmt.Println(cfg.Host)
	//fmt.Println(cfg.BearerToken)
	//fmt.Println(cfg.ServerName)
	o, err := f.OpenAPISchema()
	if err != nil {
		panic(err)
	}
	fmt.Println(o)
}
func NewFactory1() cmdutil.Factory {
	context := ""
	config := configForContext1(context)
	return cmdutil.NewFactory(config)
}
func configForContext1(context string) clientcmd.ClientConfig {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}
	if context != "" {
		overrides.CurrentContext = context
	}
	fmt.Println("--------------------------------")
	fmt.Println(overrides.ClusterDefaults.Server)
	fmt.Println("--------------------------------")
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
}
