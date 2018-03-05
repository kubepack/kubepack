package cmds

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewCmdSemver() *cobra.Command {
	var (
		minor bool
		check string
	)
	cmd := &cobra.Command{
		Use:               "semver",
		Short:             "Print sanitized semver version",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 1 {
				Fatal(errors.Errorf("multiple version found: %v", strings.Join(args, ",")))
			}
			if len(args) == 0 {
				Fatal(errors.Errorf("missing version"))
			}
			gitVersion := args[0]

			gv, err := version.NewVersion(gitVersion)
			if err != nil {
				Fatal(errors.Wrapf(err, "invalid version %s", gitVersion))
			}
			m := gv.ToMutator().ResetMetadata().ResetPrerelease()
			if minor {
				m = m.ResetPatch()
			}
			if check == "" {
				fmt.Print(m.String())
				return
			}

			c, err := version.NewConstraint(check)
			if err != nil {
				Fatal(errors.Wrapf(err, "invalid constraint %s", gitVersion))
			}
			if !c.Check(m.Done()) {
				os.Exit(1)
			}
		},
	}

	cmd.Flags().BoolVar(&minor, "minor", minor, "print major.minor.0 version")
	cmd.Flags().StringVar(&check, "check", check, "check constraint")
	return cmd
}
