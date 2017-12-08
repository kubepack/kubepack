package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
	"io/ioutil"
	"bytes"

	apps "k8s.io/api/apps/v1beta1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"github.com/ghodss/yaml"
	"strings"
)

var patchTypes = []string{"json", "merge", "strategic"}

const defaultEditor = "nano"
const _VendorFolder = "_vendor"
const PatchFolder = "patch"

var (
	srcPath   string
	patchType string
	fileInfo  os.FileInfo
)

func NewEditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit (filename)",
		Short: "Edit a file.",
		Long:  "Edit a _vendor file to generate kubectl patch.",
		Run: func(cmd *cobra.Command, args []string) {
			err := RunEdit()
			if err != nil {
				log.Println(err)
			}
		},
	}

	cmd.Flags().StringVarP(&srcPath, "src", "s", "", "File want to edit")
	cmd.Flags().StringVarP(&patchType, "type", "t", "strategic", fmt.Sprintf("Type of patch; one of %v", patchTypes))

	return cmd
}

func RunEdit() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	path := filepath.Join(root, srcPath)

	fileInfo, err = os.Stat(path)
	if err != nil {
		return err
	}

	srcFile, err := ioutil.ReadFile(path)

	buf := &bytes.Buffer{}
	buf.Write(srcFile)

	edit := NewDefaultEditor()
	edited, _, err := edit.LaunchTempFile(fmt.Sprintf("%s-edit-", filepath.Base(os.Args[0])), ".yaml", buf)

	srcJson, err := yaml.YAMLToJSON(srcFile)
	if err != nil {
		return err
	}

	dstJson, err := yaml.YAMLToJSON(edited)
	if err != nil {
		return err
	}

	return GetPatch(srcJson, dstJson)
}

func GetPatch(src, dst []byte) error {
	var err error
	var patch []byte

	switch patchType {
	case "strategic":
		patch, err = strategicpatch.CreateTwoWayMergePatch(src, dst, apps.Deployment{})
	}
	if err != nil {
		return err
	}
	yamlPatch, err := yaml.JSONToYAML(patch)
	if err != nil {
		return err
	}

	root, err := os.Getwd()
	if err != nil {
		return err
	}
	patchFolderDir := strings.Replace(srcPath, _VendorFolder, PatchFolder, 1)
	lstIndexSlash := strings.LastIndex(patchFolderDir, "/")
	dstPath := filepath.Join(root, patchFolderDir[0:lstIndexSlash])

	err = os.MkdirAll(dstPath, 0755)
	if err != nil {
		return err
	}
	patchFilePath := filepath.Join(dstPath, fileInfo.Name())
	_, err = os.Create(patchFilePath)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(patchFilePath, yamlPatch, 0755)
	if err != nil {
		return err
	}

	return nil
}

func NewDefaultEditor() editor.Editor {
	return editor.Editor{
		Args:  []string{defaultEditor},
		Shell: false,
	}
}
