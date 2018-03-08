package commands

import (
	"log"
	"os"
	"path/filepath"

	api "github.com/kubepack/pack-server/apis/manifest/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewKubepackInitializeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize kubepack and create manifest.yaml file",
		Run: func(cmd *cobra.Command, args []string) {
			err := createManifestFile(cmd)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	return cmd
}

func createManifestFile(cmd *cobra.Command) error {
	wd, err := cmd.Flags().GetString("file")
	if err != nil {
		errors.WithStack(err)
	}
	mPath := filepath.Join(wd, api.DependencyFile)
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
