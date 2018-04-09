package commands

import (
	"github.com/spf13/cobra"
	"log"
	"github.com/pkg/errors"
	"path/filepath"
	"os"
	api "github.com/kubepack/pack-server/apis/manifest/v1alpha1"
	"fmt"
)

type deleteOptions struct{}

const DeleteShScriptName = "delete.sh"

func NewDeleteCommand(plugin bool) *cobra.Command {
	var d deleteOptions
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Generate delete script for removing deployed instances.",
		Run: func(cmd *cobra.Command, args []string) {
			err := d.Validate(cmd, args)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}

	return cmd
}

func (d deleteOptions) Validate(cmd *cobra.Command, args []string) error {
	root, err := cmd.Flags().GetString("file")
	if err != nil {
		return errors.WithStack(err)
	}
	if root == "" {
		return errors.Errorf("Needs to provide filepath")
	}
	return nil
}

func (d deleteOptions) generateDeleteShScript(cmd *cobra.Command, plugin bool) error {
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

	return nil
}

func generateDeleteDag(root string) error {
	var check map[string]int
	check = make(map[string]int)
	deleteShPath := filepath.Join(root, api.ManifestDirectory, CompileDirectory, DeleteShScriptName)
	fmt.Println(deleteShPath)
	if _, err := os.Stat(deleteShPath); err == nil {
		err = os.Remove(deleteShPath)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	err := os.MkdirAll(filepath.Dir(deleteShPath), 0755)
	if err != nil {
		return errors.WithStack(err)
	}
	err = WriteCompiledFileToDest(deleteShPath, []byte(InstallSHDefault))
	if err != nil {
		return errors.WithStack(err)
	}
	f, err := os.OpenFile(deleteShPath, os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	manifestPath := filepath.Join(root, api.DependencyFile)
	manVendorDir := filepath.Join(root, api.ManifestDirectory, _VendorFolder)
	depList, err := getManifestStruct(manifestPath)
	if err != nil {
		return errors.WithStack(err)
	}
	st := NewStack()
	// check
	for _, val := range depList.Items {
		st.Push(val.Package)
		check[val.Package] = 1
	}
	for _, val := range depList.Items {
		st.Push(val.Package)
		check[val.Package] = 1
	}
	for len(st.s) > 0 {
		n, err := st.Pop()
		if err != nil {
			return errors.WithStack(err)
		}
		manifestPath = filepath.Join(manVendorDir, n, api.DependencyFile)
		data, err := getManifestStruct(manifestPath)
		if err != nil {
			return errors.WithStack(err)
		}
		for _, val := range data.Items {
			if _, ok := check[val.Package]; !ok {
				st.Push(val.Package)
			}

			check[val.Package] = max(check[n]+1, check[val.Package])
		}
	}
	return nil
}
