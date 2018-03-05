package cmds

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/drone/envsubst"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewCmdEnvsubst() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "envsubst",
		Short:             "Emulates bash environment variable substitution for input text",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(os.Stdin)
			data, err := ioutil.ReadAll(reader)
			if err != nil {
				Fatal(errors.Wrap(err, "failed to read input"))
			}
			out, err := envsubst.EvalEnv(string(data))
			if err != nil {
				Fatal(errors.Wrap(err, "failed to decode input"))
			}
			fmt.Print(string(out))
		},
	}
	return cmd
}
