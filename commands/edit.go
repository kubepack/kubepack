package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanphx/json-patch"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
	"k8s.io/kubernetes/pkg/kubectl/scheme"
)

const defaultEditor = "nano"
const _VendorFolder = "vendor"
const PatchFolder = "patch"

var (
	srcPath  string
	fileInfo os.FileInfo
)

func NewEditCommand(plugin bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit (filename)",
		Short: "Edit resource definition",
		Long:  "Generates patch via edit command",

		Run: func(cmd *cobra.Command, args []string) {
			var err error
			srcPath, err = cmd.Flags().GetString("src")
			if err != nil {
				log.Println(err)
			}
			err = RunEdit(cmd)
			if err != nil {
				log.Println(err)
			}
		},
	}

	return cmd
}

func RunEdit(cmd *cobra.Command) error {
	root, err := cmd.Flags().GetString("file")
	if err != nil {
		return errors.WithStack(err)
	}
	path := filepath.Join(root, srcPath)

	fileInfo, err = os.Stat(path)
	if err != nil {
		return errors.WithStack(err)
	}

	srcFile, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.WithStack(err)
	}

	buf := &bytes.Buffer{}
	buf.Write(srcFile)

	edit := NewDefaultEditor()
	edited, _, err := edit.LaunchTempFile(fmt.Sprintf("%s-edit-", filepath.Base(os.Args[0])), ".yaml", buf)

	srcJson, err := yaml.YAMLToJSON(srcFile)
	if err != nil {
		return errors.WithStack(err)
	}

	dstJson, err := yaml.YAMLToJSON(edited)
	if err != nil {
		return errors.WithStack(err)
	}

	return GetPatch(srcJson, dstJson, cmd)
}

func GetPatch(src, dst []byte, cmd *cobra.Command) error {
	var err error
	var patch []byte

	// ref: https://github.com/kubernetes/kubernetes/blob/master/pkg/kubectl/cmd/util/editor/editoptions.go#L549

	var ro runtime.TypeMeta
	if err := yaml.Unmarshal(src, &ro); err != nil {
		return errors.WithStack(err)
	}
	kind := ro.GetObjectKind().GroupVersionKind()
	versionedObject, err := scheme.Scheme.New(kind)

	switch {
	case runtime.IsNotRegisteredError(err):
		patch, err = jsonpatch.CreateMergePatch(src, dst)
	case err != nil:
		return errors.WithStack(err)
	default:
		patch, err = strategicpatch.CreateTwoWayMergePatch(src, dst, versionedObject)
	}

	yamlPatch, err := yaml.JSONToYAML(patch)
	if err != nil {
		return errors.WithStack(err)
	}
	root, err := cmd.Flags().GetString("file")
	if err != nil {
		return errors.WithStack(err)
	}
	patchFolderDir := strings.Replace(srcPath, _VendorFolder, PatchFolder, 1)
	lstIndexSlash := strings.LastIndex(patchFolderDir, "/")
	dstPath := filepath.Join(root, patchFolderDir[0:lstIndexSlash])

	err = os.MkdirAll(dstPath, 0755)
	if err != nil {
		return errors.WithStack(err)
	}
	patchFilePath := filepath.Join(dstPath, fileInfo.Name())
	_, err = os.Create(patchFilePath)
	if err != nil {
		return errors.WithStack(err)
	}
	err = ioutil.WriteFile(patchFilePath, yamlPatch, 0755)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func NewDefaultEditor() editor.Editor {
	return editor.Editor{
		Args:  []string{defaultEditor},
		Shell: false,
	}
}
