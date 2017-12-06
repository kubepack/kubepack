package cmd

import (
	"github.com/spf13/cobra"
	"github.com/evanphx/json-patch"
	apps "k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/Masterminds/vcs"
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
		return nil, err
	}

	srcDir := filepath.Join(root, src)
	_, err = os.Stat(srcDir)
	if err != nil {
		return nil, err
	}

	srcFile, err := ioutil.ReadFile(srcDir)
	if err != nil {
		return nil, err
	}

	jsonSrc, err := yaml.YAMLToJSON(srcFile)
	if err != nil {
		return nil, err
	}

	patchDir := filepath.Join(root, patch)
	patchFileInfo, err = os.Stat(patchDir)
	if err != nil {
		return nil, err
	}

	patchFile, err := ioutil.ReadFile(patchDir)
	if err != nil {
		return nil, err
	}

	jsonPatch, err := yaml.YAMLToJSON(patchFile)
	if err != nil {
		return nil, err
	}

	compiled, err := jsonpatch.MergePatch(jsonSrc, jsonPatch)
	if err != nil {
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
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	annotateYaml, err := getAnnotatedWithCommitHash(compiledYaml, root)
	if err != nil {
		return err
	}

	outlookDir := strings.Replace(patch, PatchFolder, CompileDirectory, 1)
	lstIndexOfSlash := strings.LastIndex(outlookDir, "/")
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

	err = ioutil.WriteFile(outLookFilePath, annotateYaml, 0755)
	if err != nil {
		return err
	}

	return nil
}

func getAnnotatedWithCommitHash(yamlByte []byte, dir string) ([]byte, error) {
	repo, err := vcs.NewRepo("", dir)
	if err != nil {
		return nil, err
	}

	crnt, err := repo.Current()
	if err != nil {
		return nil, err
	}

	commitInfo, err := repo.CommitInfo(string(crnt))
	if err != nil {
		return nil, err
	}

	deploy := &apps.Deployment{}
	err = yaml.Unmarshal(yamlByte, deploy)
	metav1.SetMetaDataAnnotation(&deploy.ObjectMeta, "git-commit-hash", commitInfo.Commit)

	annotatedYamlByte, err := yaml.Marshal(deploy)
	if err != nil {
		return nil, err
	}

	fmt.Println("Hello world---------------After annotated yaml---", string(annotatedYamlByte))

	return annotatedYamlByte, nil
}
