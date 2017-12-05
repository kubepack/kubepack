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
)

var (
	src   string
	patch string
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

func DumpCompiledFile (compiledYaml []byte)  {
	fmt.Println("hello yaml", string(compiledYaml))
}
