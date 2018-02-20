package cmds

import (
	"github.com/spf13/cobra"
	"os"
	"github.com/pkg/errors"
	"path/filepath"
	"github.com/kubepack/kubepack/type"
	"log"
)

func NewKubepackInitializeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize kubepack and create manifest.yaml file",
		Run: func(cmd *cobra.Command, args []string) {
			err := createManifestFile()
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	return cmd
}

func createManifestFile() error {
	wd, err := os.Getwd()
	if err != nil {
		errors.WithStack(err)
	}
	mPath := filepath.Join(wd, types.ManifestFile)
	if _, err := os.Stat(mPath); err != nil {
		if os.IsNotExist(err) {
			_, err = os.Create(mPath)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return errors.WithStack(err)
	}
	return nil
}
