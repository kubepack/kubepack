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
	// "k8s.io/apimachinery/pkg/util/jsonmergepatch"
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
			RunEdit()
		},
	}

	cmd.Flags().StringVarP(&srcPath, "src", "s", "", "File want to edit")
	cmd.Flags().StringVarP(&patchType, "type", "t", "strategic", fmt.Sprintf("Type of patch; one of %v", patchTypes))

	return cmd
}

func AddEditFlag(cmd *cobra.Command) {
	// cmd.Flags().StringP("file", "f", "888888", "Edit file, provided through the --file flag.")
}

func RunEdit() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatalln("Error during get root path.", err)
	}
	path := filepath.Join(root, srcPath)

	fileInfo, err = os.Stat(path)
	if err != nil {
		log.Fatalln("--------------", err)
	}

	srcFile, err := ioutil.ReadFile(path)

	buf := &bytes.Buffer{}
	buf.Write(srcFile)

	edit := NewDefaultEditor()
	edited, _, err := edit.LaunchTempFile(fmt.Sprintf("%s-edit-", filepath.Base(os.Args[0])), ".yaml", buf)

	srcJson, err := yaml.YAMLToJSON(srcFile)
	if err != nil {
		log.Fatalln("Error, while converting source from yaml to json", err)
	}

	dstJson, err := yaml.YAMLToJSON(edited)
	if err != nil {
		log.Fatalln("Error, while converting destination from yaml to json", err)
	}

	GetPatch(srcJson, dstJson)
}

func GetPatch(src, dst []byte) {
	var err error
	var patch []byte

	switch patchType {
	case "strategic":
		patch, err = strategicpatch.CreateTwoWayMergePatch(src, dst, apps.Deployment{})
	}
	if err != nil {
		log.Fatalln("Error to generate patch", err)
	}
	yamlPatch, err := yaml.JSONToYAML(patch)
	if err != nil {
		log.Fatalln(err)
	}
	// strings.LastIndex(strings.Replace(srcPath, _VendorFolder, PatchFolder, 1), "/")
	root, err := os.Getwd()
	if err != nil {
		log.Fatalln("Error during get root path.", err)
	}
	lstIndexSlash := strings.LastIndex(strings.Replace(srcPath, _VendorFolder, PatchFolder, 1), "/")
	dstPath := filepath.Join(root, strings.Replace(srcPath, _VendorFolder, PatchFolder, 1)[0:lstIndexSlash])
	fmt.Println("Patch: ", dstPath)

	err = os.MkdirAll(dstPath, 0755)
	if err != nil {
		log.Fatalln("Error to MkdirAll", err)
	}
	patchFilePath := filepath.Join(dstPath, fileInfo.Name())
	_, err = os.Create(patchFilePath)
	if err != nil {
		log.Fatalln("Error creating patch file.", err)
	}

	err = ioutil.WriteFile(patchFilePath, yamlPatch, 07555)
	if err != nil {
		log.Fatalln("Error to writing patch file.", err)
	}

	// os.Open()

	fmt.Println(string(yamlPatch))

}

func NewDefaultEditor() editor.Editor {
	return editor.Editor{
		Args:  []string{defaultEditor},
		Shell: false,
	}
}
