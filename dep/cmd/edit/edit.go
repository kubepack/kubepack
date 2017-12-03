package edit

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
)

var patchTypes = []string{"json", "merge", "strategic"}

const defaultEditor = "nano"

var (
	srcPath   string
	patchType string
)

func NewEditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit (filename)",
		Short: "Edit a file.",
		Long:  "Edit a _vendor file to generate kubectl patch.",
		Run: func(cmd *cobra.Command, args []string) {
			RunEdit(cmd)
		},
	}

	cmd.Flags().StringVarP(&srcPath, "src", "s", "", "File want to edit")
	cmd.Flags().StringVarP(&patchType, "type", "t", "strategic", fmt.Sprintf("Type of patch; one of %v", patchTypes))

	return cmd
}

func AddEditFlag(cmd *cobra.Command) {
	// cmd.Flags().StringP("file", "f", "888888", "Edit file, provided through the --file flag.")
}

func RunEdit(cmd *cobra.Command) {
	root, err := os.Getwd()
	if err != nil {
		fmt.Errorf("Error during get root path.", err)
	}
	path := filepath.Join(root, srcPath)

	_, err = os.Stat(path)
	if err != nil {
		fmt.Errorf("--------------", err)
	}

	srcFile, err := ioutil.ReadFile(path)

	buf := &bytes.Buffer{}
	buf.Write(srcFile)

	edit := NewDefaultEditor()
	edited, _, err := edit.LaunchTempFile(fmt.Sprintf("%s-edit-", filepath.Base(os.Args[0])), ".yaml", buf)

	srcJson, err := yaml.YAMLToJSON(srcFile)
	if err != nil {
		fmt.Errorf("Error, while converting source from yaml to json", err)
	}

	dstJson, err := yaml.YAMLToJSON(edited)
	if err != nil {
		fmt.Errorf("Error, while converting destination from yaml to json", err)
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
		fmt.Errorf("Error to generate patch", err)
	}
	yamlPatch, err := yaml.JSONToYAML(patch)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Patch: ")
	fmt.Println(string(yamlPatch))
}

func NewDefaultEditor() editor.Editor {
	return editor.Editor{
		Args:  []string{defaultEditor},
		Shell: false,
	}
}
