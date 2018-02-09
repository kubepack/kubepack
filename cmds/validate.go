package cmds

import "github.com/spf13/cobra"
import (
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"fmt"
)

func NewValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate _outlook folder",
		Run: func(cmd *cobra.Command, args []string) {
			err := validateOutlook()
			if err != nil {
				panic(err)
			}
		},
	}
	return cmd
}

func validateOutlook() error {
	f := cmdutil.NewFactory("/home/tigerworks/.kube/config")
	// schema, err := f.Validator(true)
	/*if err != nil {
		panic(err)
	}*/

	openapiSchema, err := f.OpenAPISchema()
	if err != nil {
		panic(err)
	}

	fmt.Println(openapiSchema)


	return nil
}
