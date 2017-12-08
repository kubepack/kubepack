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
			rootPath, err := os.Getwd()
			if err != nil {
				log.Fatalln(err)
			}

			err = filepath.Walk(filepath.Join(rootPath, PatchFolder), visitPatchAndDump)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}

	cmd.Flags().StringVarP(&src, "src", "s", src, "Compile patch and source.")
	cmd.Flags().StringVarP(&patch, "patch", "p", patch, "Compile patch and source.")

	return cmd
}

func visitPatchAndDump(path string, fileInfo os.FileInfo, err error) error {
	if fileInfo.IsDir() {
		return nil
	}
	patchByte, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	srcFilepath := strings.Replace(path, PatchFolder, _VendorFolder, 1)

	if _, err := os.Stat(srcFilepath); err != nil {
		return err
	}

	srcYamlByte, err := ioutil.ReadFile(srcFilepath)
	if err != nil {
		return err
	}

	mergedPatchYaml, err := CompileWithPatch(srcYamlByte, patchByte)
	if err != nil {
		return err
	}

	err = DumpCompiledFile(mergedPatchYaml, strings.Replace(path, PatchFolder, CompileDirectory, 1))
	if err != nil {
		return err
	}
	return nil
}

func CompileWithPatch(srcByte, patchByte []byte) ([]byte, error) {
	jsonSrc, err := yaml.YAMLToJSON(srcByte)
	if err != nil {
		return nil, err
	}

	jsonPatch, err := yaml.YAMLToJSON(patchByte)
	if err != nil {
		return nil, err
	}

	compiled, err := jsonpatch.MergePatch(jsonSrc, jsonPatch)
	if err != nil {
		return nil, err
	}

	yaml, err := yaml.JSONToYAML(compiled)
	if err != nil {
		return nil, err
	}
	return yaml, err
}

func DumpCompiledFile(compiledYaml []byte, outlookPath string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	annotateYaml, err := getAnnotatedWithCommitHash(compiledYaml, root)
	if err != nil {
		return err
	}

	// If not exists mkdir all the folder
	outlookDir := filepath.Dir(outlookPath)
	if _, err := os.Stat(outlookDir); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(outlookDir, 0755)
			if err != nil {
				return err
			}
		}
	}

	_, err = os.Create(outlookPath)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(outlookPath, annotateYaml, 0755)
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

	return annotatedYamlByte, nil
}
