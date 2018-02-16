package cmds

import (
	"io/ioutil"
	"github.com/appscode/go/log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/vcs"
	"github.com/evanphx/json-patch"
	"github.com/ghodss/yaml"
	typ "github.com/kubepack/kubepack/type"
	"github.com/spf13/cobra"
	"github.com/pkg/errors"
	"github.com/google/go-jsonnet"
)

var (
	src   string
	patch string
)

const CompileDirectory = "_outlook"

// var validator *validation.SchemaValidation

func NewUpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Compiles patches and vendored manifests into final resource definitions",
		Run: func(cmd *cobra.Command, args []string) {
			rootPath, err := os.Getwd()
			if err != nil {
				log.Fatalln(err)
			}
			validator, err = GetOpenapiValidator(cmd)
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
	if fileInfo.Name() == ".gitignore" || fileInfo.Name() == "README.md" {
		return nil
	}
	if strings.HasSuffix(path, "jsonnet.TEMPLATE") {
		return nil
	}
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
		err = validator.ValidateBytes(srcYamlByte)
		if err != nil {
			vm := jsonnet.MakeVM()
			j, err := vm.EvaluateSnippet(path, string(srcYamlByte))
			if err != nil {
				return errors.Wrap(err, "Error to evaluate jsonet")
			}
			yml, err := yaml.JSONToYAML([]byte(j))
			if err != nil {
				errors.Wrap(err, "error to convert json to yaml")
			}
			srcYamlByte = yml
		}
		err = DumpCompiledFile(srcYamlByte, strings.Replace(path, _VendorFolder, CompileDirectory, 1))
		if err != nil {
			return errors.Wrap(err, "Error dump compiled file")
		}
		return nil
	}

	patchByte, err := ioutil.ReadFile(strings.Replace(path, _VendorFolder, PatchFolder, 1))
	if err != nil {
		return errors.Wrap(err, "Error to read patch file")
	}

	splitWithVendor := strings.Split(path, _VendorFolder)
	if len(splitWithVendor) != 2 {
		return nil
	}

	mergedPatchYaml, err := CompileWithPatch(srcYamlByte, patchByte)
	if err != nil {
		return errors.Wrap(err, "Error to merge patch")
	}

	err = DumpCompiledFile(mergedPatchYaml, strings.Replace(path, _VendorFolder, CompileDirectory, 1))
	if err != nil {
		return errors.Wrap(err, "Error to dump compiled file")
	}
	return nil
}

func CompileWithPatch(srcByte, patchByte []byte) ([]byte, error) {
	jsonSrc, err := yaml.YAMLToJSON(srcByte)
	if err != nil {
		return nil, errors.Wrap(err, "Error to convert source yaml to json.")
	}

	jsonPatch, err := yaml.YAMLToJSON(patchByte)
	if err != nil {
		return nil, errors.Wrap(err, "Error to convert patch yaml to json.")
	}

	compiled, err := jsonpatch.MergePatch(jsonSrc, jsonPatch)
	if err != nil {
		return nil, errors.Wrap(err, "Error to marge patch with source.")
	}

	compiledYaml, err := yaml.JSONToYAML(compiled)
	if err != nil {
		return nil, errors.Wrap(err, "Error to convert compiled yaml to json.")
	}
	return compiledYaml, nil
}

func DumpCompiledFile(compiledYaml []byte, outlookPath string) error {
	root, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Error to get wd(os.Getwd()).")
	}
	annotateYaml, err := getAnnotatedWithCommitHash(compiledYaml, root)
	if err != nil {
		return errors.Wrap(err, "error to annotated with git-commit-hash")
	}

	// If not exists mkdir all the folder
	outlookDir := filepath.Dir(outlookPath)
	if _, err := os.Stat(outlookDir); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(outlookDir, 0755)
			if err != nil {
				return errors.Wrap(err, "Error to mkdir.")
			}
		}
	}

	_, err = os.Create(outlookPath)
	if err != nil {
		return errors.Wrap(err, "Error to create outlook.")
	}

	err = ioutil.WriteFile(outlookPath, annotateYaml, 0755)
	if err != nil {
		return errors.Wrap(err, "Error to write file in outlook folder.")
	}

	return nil
}

func getAnnotatedWithCommitHash(yamlByte []byte, dir string) ([]byte, error) {
	repo, err := getRootDir(dir)
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

	annotatedMap := map[string]interface{}{}
	err = yaml.Unmarshal(yamlByte, &annotatedMap)
	if err != nil {
		return nil, err
	}

	if err != nil {
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
	for {
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
