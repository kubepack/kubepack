package cmds

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/vcs"
	"github.com/evanphx/json-patch"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	typ "github.com/kubepack/pack/type"
	"k8s.io/kubernetes/pkg/api"
)

var (
	src   string
	patch string
)

const CompileDirectory = "_outlook"

func NewUpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Compiles patches and vendored manifests into final resource definitions",
		Run: func(cmd *cobra.Command, args []string) {
			rootPath, err := os.Getwd()
			if err != nil {
				log.Fatalln(err)
			}

			err = filepath.Walk(filepath.Join(rootPath, _VendorFolder), visitPatchAndDump)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}

	cmd.Flags().StringVarP(&src, "src", "s", src, "Compile patch and source.")
	cmd.Flags().StringVarP(&patch, "patch", "p", patch, "Compile patch and source.")

	return cmd
}

func visitPatchAndDump(path string, fileInfo os.FileInfo, ferr error) error {
	if ferr != nil {
		return ferr
	}

	if fileInfo.IsDir() {
		return nil
	}

	if fileInfo.Name() == typ.ManifestFile {
		return nil
	}

	srcFilepath := path
	srcYamlByte, err := ioutil.ReadFile(srcFilepath)
	if err != nil {
		return err
	}

	patchFilePath := strings.Replace(path, _VendorFolder, PatchFolder, 1)
	if _, err := os.Stat(patchFilePath); err != nil {
		err = DumpCompiledFile(srcYamlByte, strings.Replace(path, _VendorFolder, CompileDirectory, 1))
		if err != nil {
			return err
		}
		return nil
	}

	patchByte, err := ioutil.ReadFile(strings.Replace(path, _VendorFolder, PatchFolder, 1))
	if err != nil {
		return err
	}

	splitWithVendor := strings.Split(path, _VendorFolder)
	if len(splitWithVendor) != 2 {
		return nil
	}

	mergedPatchYaml, err := CompileWithPatch(srcYamlByte, patchByte)
	if err != nil {
		return err
	}

	err = DumpCompiledFile(mergedPatchYaml, strings.Replace(path, _VendorFolder, CompileDirectory, 1))
	if err != nil {
		return err
	}
	return nil
}

func getVersionedObject(json []byte) (runtime.Object, error) {
	var ro runtime.TypeMeta
	if err := yaml.Unmarshal(json, &ro); err != nil {
		return nil, err
	}
	kind := ro.GetObjectKind().GroupVersionKind()
	return api.Scheme.New(kind)
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

	compiledYaml, err := yaml.JSONToYAML(compiled)
	if err != nil {
		return nil, err
	}
	return compiledYaml, err
}

func DumpCompiledFile(compiledYaml []byte, outlookPath string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	annotateYaml, err := getAnnotatedWithCommitHash(compiledYaml, root)
	if err != nil {
		fmt.Println(err)
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
	repo, err := getRootDir(dir)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	crnt, err := repo.Current()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	commitInfo, err := repo.CommitInfo(string(crnt))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	annotatedMap := map[string]interface{}{}
	err = yaml.Unmarshal(yamlByte, &annotatedMap)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	metadata := annotatedMap["metadata"]
	annotations, ok := metadata.(map[string]interface{})["annotations"]
	if !ok || annotations == nil {
		metadata.(map[string]interface{})["annotations"] = map[string]interface{}{}
		annotations = metadata.(map[string]interface{})["annotations"]
	}
	annotations.(map[string]interface{})["git-commit-hash"] = commitInfo.Commit
	annotatedMap["metadata"] = metadata

	return yaml.Marshal(annotatedMap)
}


func getRootDir(path string) (vcs.Repo, error) {
	var err error
	for ; ; {
		repo, err := vcs.NewRepo("", path)
		if err == nil {
			return repo, err
		}
		if os.Getenv("HOME") == path {
			break
		}
		path = filepath.Dir(path)
	}

	return nil, err
}
