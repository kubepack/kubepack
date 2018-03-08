package commands

import (
	"log"
	"os"
	"path/filepath"

	api "github.com/kubepack/pack-server/apis/manifest/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewKubepackInitializeCmd(plugin bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize kubepack and create manifest.yaml file",
		Run: func(cmd *cobra.Command, args []string) {
			err := createManifestFile(cmd, plugin)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	return cmd
}

func createManifestFile(cmd *cobra.Command, plugin bool) error {
	root, err := cmd.Flags().GetString("file")
	if err != nil {
		errors.WithStack(err)
	}
	if !plugin && !filepath.IsAbs(root) {
		wd, err := os.Getwd()
		if err != nil {
			return errors.WithStack(err)
		}
		root = filepath.Join(wd, root)
	}
	if !filepath.IsAbs(root) {
		return errors.Errorf("Need to provide Absolute path. Here is the issue: https://github.com/kubernetes/kubectl/issues/346")
	}
	mPath := filepath.Join(root, api.DependencyFile)
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
