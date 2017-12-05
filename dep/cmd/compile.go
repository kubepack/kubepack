package cmd

import (
	"github.com/spf13/cobra"
	"github.com/evanphx/json-patch"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
	"log"
	"path/filepath"
	"fmt"
	"strings"
)

var (
	src           string
	patch         string
	patchFileInfo os.FileInfo
)

const CompileDirectory = "_outlook"

func NewCompileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compile",
		Short: "Compile with patch.",
		Run: func(cmd *cobra.Command, args []string) {
			compiledYaml, err := CompileWithPatch()
			if err != nil {
				log.Fatalln(err)
			}
			// fmt.Println("yaml", string(compiledYaml))
			DumpCompiledFile(compiledYaml)
		},
	}

	cmd.Flags().StringVarP(&src, "src", "s", src, "Compile patch and source.")
	cmd.Flags().StringVarP(&patch, "patch", "p", patch, "Compile patch and source.")

	return cmd
}

func CompileWithPatch() ([]byte, error) {
	root, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	srcDir := filepath.Join(root, src)
	_, err = os.Stat(srcDir)
	if err != nil {
		return nil, err
	}

	srcFile, err := ioutil.ReadFile(srcDir)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	jsonSrc, err := yaml.YAMLToJSON(srcFile)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	patchDir := filepath.Join(root, patch)
	patchFileInfo, err = os.Stat(patchDir)
	if err != nil {
		return nil, err
	}

	patchFile, err := ioutil.ReadFile(patchDir)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	jsonPatch, err := yaml.YAMLToJSON(patchFile)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	compiled, err := jsonpatch.MergePatch(jsonSrc, jsonPatch)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	yaml, err := yaml.JSONToYAML(compiled)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	fmt.Println(string(yaml))
	return yaml, err
}

func DumpCompiledFile(compiledYaml []byte) error {
	fmt.Println("hello yaml", string(compiledYaml))
	outlookDir := strings.Replace(patch, PatchFolder, CompileDirectory, 1)
	fmt.Println("path")
	lstIndexOfSlash := strings.LastIndex(outlookDir, "/")
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	dstPath := filepath.Join(root, outlookDir[0:lstIndexOfSlash])
	err = os.MkdirAll(dstPath, 0755)
	if err != nil {
		return err
	}
	outLookFilePath := filepath.Join(dstPath, patchFileInfo.Name())
	fmt.Println("file name-----", outLookFilePath)
	_, err = os.Create(outLookFilePath)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(outLookFilePath, compiledYaml, 0755)
	if err != nil {
		return err
	}

	return nil
}
