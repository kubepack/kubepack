package cmds

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewCmdBase64() *cobra.Command {
	var (
		decode bool
	)
	cmd := &cobra.Command{
		Use:               "base64",
		Short:             "Base64 encode/decode input text",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(os.Stdin)
			data, err := ioutil.ReadAll(reader)
			if err != nil {
				Fatal(errors.Wrap(err, "failed to read input"))
			}
			if decode {
				out, err := base64.StdEncoding.DecodeString(string(data))
				if err != nil {
					Fatal(errors.Wrap(err, "failed to decode input"))
				}
				fmt.Print(string(out))
			} else {
				fmt.Print(base64.StdEncoding.EncodeToString(data))
			}
		},
	}

	cmd.Flags().BoolVar(&decode, "decode", decode, "Decode input text")
	return cmd
}
