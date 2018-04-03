package commands

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	api "github.com/kubepack/pack-server/apis/manifest/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	kin_const "k8s.io/kubectl/pkg/kinflate/constants"
)

type initOptions struct{}

const manifestTemplate = `apiVersion: manifest.k8s.io/v1alpha1
kind: Manifest
metadata:
  name: helloworld
description: helloworld does useful stuff.
namePrefix: some-prefix
# Labels to add to all objects and selectors.
# These labels would also be used to form the selector for apply --prune
# Named differently than “labels” to avoid confusion with metadata for this object
objectLabels:
  app: helloworld
objectAnnotations:
  note: This is an example annotation
# Add Resources to inflate below. List of directories/file-paths to add.
resources: []
#- service.yaml
#- ../some-dir/
# There could also be configmaps in Base, which would make these overlays
configmaps: []
# There could be secrets in Base, if just using a fork/rebase workflow
secrets: []
recursive: true
`

func NewKubepackInitializeCmd(plugin bool) *cobra.Command {
	var o initOptions
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize kubepack and create manifest.yaml file",
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Validate(cmd, args)
			if err != nil {
				log.Fatalln(err)
			}
			err = createManifestFiles(cmd, plugin)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	return cmd
}

func (o *initOptions) Validate(cmd *cobra.Command, args []string) error {
	root, err := cmd.Flags().GetString("file")
	if err != nil {
		return errors.WithStack(err)
	}
	if root == "" {
		return errors.Errorf("Needs to provide filepath")
	}
	return nil
}

func createManifestFiles(cmd *cobra.Command, plugin bool) error {
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
		return errors.Errorf("Duh! we need an absolute path when used as a kubectl plugin. For more info, see here: https://github.com/kubernetes/kubectl/issues/346")
	}
	mPath := filepath.Join(root, api.DependencyFile)
	_, err = os.Stat(mPath)
	if err == nil {
		err = initializeKubeManifests(root)
		if err != nil {
			return errors.Wrapf(err, "%s already exists.", api.DependencyFile)
		}
		return errors.Errorf("%s already exists.", api.DependencyFile)
	}
	if err != nil {
		if os.IsNotExist(err) {
			_, err = os.Create(mPath)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}
	err = initializeKubeManifests(root)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func initializeKubeManifests(root string) error {
	mPath := filepath.Join(root, kin_const.KubeManifestFileName)
	_, err := os.Stat(mPath)
	if err == nil {
		return errors.Errorf("%s already exists.", kin_const.KubeManifestFileName)
	}

	return ioutil.WriteFile(mPath, []byte(manifestTemplate), 0666)
}
