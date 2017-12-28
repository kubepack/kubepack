package cmds

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/api"
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
		Short: "Edit resource definition",
		Long:  "Generates patch via edit command",
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

	var ro runtime.TypeMeta
	if err := yaml.Unmarshal(src, &ro); err != nil {
		return err
	}
	kind := ro.GetObjectKind().GroupVersionKind()
	versionedObject, err := api.Scheme.New(kind)

	switch {
	case runtime.IsNotRegisteredError(err):
		patch, err = jsonpatch.CreateMergePatch(src, dst)
	default:
		patch, err = strategicpatch.CreateTwoWayMergePatch(src, dst, versionedObject)
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
