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

const DeleteShScriptName = "uninstall.sh"

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
	dag, err := generateDAG(root, DeleteShScriptName)
	if err != nil {
		return errors.WithStack(err)
	}
	deleteShPath := filepath.Join(root, api.ManifestDirectory, CompileDirectory, DeleteShScriptName)
	if _, err := os.Stat(deleteShPath); err == nil {
		err = os.Remove(deleteShPath)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	err = os.MkdirAll(filepath.Dir(deleteShPath), 0755)
	if err != nil {
		return errors.WithStack(err)
	}
	err = WriteCompiledFileToDest(deleteShPath, []byte(InstallSHDefault))
	if err != nil {
		return errors.WithStack(err)
	}
	dag = append(dag, node{GetImportRoot(root), 0})
	sort.Slice(dag, func(i, j int) bool {
		return dag[i].count < dag[j].count
	})
	for _, p := range dag {
		err = writeCommandToDeleteSH(p.node, root)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	uninstallPath := filepath.Join(rootPath, api.ManifestDirectory, CompileDirectory, DeleteShScriptName)
	err = os.Chmod(uninstallPath, 0777)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func writeCommandToDeleteSH(pkg, root string) error {
	deleteTemplate := `
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
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return nil
	}
	cmd := `kubectl delete -R -f .`
	installShContent := fmt.Sprintf(deleteTemplate, outputPath, cmd)
	_, err = f.Write([]byte(installShContent))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
