package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	api "github.com/kubepack/pack-server/apis/manifest/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
			err = d.generateDeleteShScript(cmd, plugin)
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
	err = generateDeleteDag(root)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func generateDeleteDag(root string) error {
	var res []node
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
		fmt.Println(nil)
		fmt.Println(data)
		fmt.Println(err)
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
	for key, val := range check {
		n := node{
			node:  key,
			count: val,
		}
		res = append(res, n)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].count < res[j].count
	})
	for _, val := range res {
		err = writeCommandToInstallSH(val.node, root)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	fmt.Println(check)
	return nil
}

func writeCommandToDeleteSH(pkg, root string) error {
	installTemplate := `
pushd %s
%s
popd
			
`
	installPath := filepath.Join(root, api.ManifestDirectory, CompileDirectory, DeleteShScriptName)
	f, err := os.OpenFile(installPath, os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	outputPath := filepath.Join(api.ManifestDirectory, CompileDirectory, pkg)
	cmd := ``
	installShContent := fmt.Sprintf(installTemplate, outputPath, cmd)
	_, err = f.Write([]byte(installShContent))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
