package commands

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/vcs"
	"github.com/appscode/go/log"
	"github.com/evanphx/json-patch"
	"github.com/ghodss/yaml"
	"github.com/google/go-jsonnet"
	api "github.com/kubepack/pack-server/apis/manifest/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/kubectl/pkg/kinflate/resource"
	"k8s.io/kubernetes/pkg/kubectl/scheme"
)

var (
	src        string
	patch      string
	rootPath   string
	patchFiles map[string]string
)

const CompileDirectory = "output"

func NewUpCommand(plugin bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Compiles patches and vendored manifests into final resource definitions",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			rootPath, err = cmd.Flags().GetString("file")
			if err != nil {
				log.Fatalln(errors.WithStack(err))
			}
			if !plugin && !filepath.IsAbs(rootPath) {
				wd, err := os.Getwd()
				if err != nil {
					log.Fatalln(errors.WithStack(err))
				}
				rootPath = filepath.Join(wd, rootPath)
			}
			if !filepath.IsAbs(rootPath) {
				log.Fatalln(errors.Errorf("Duh! we need an absolute path when used as a kubectl plugin. For more info, see here: https://github.com/kubernetes/kubectl/issues/346"))
			}
			validator, err = GetOpenapiValidator(cmd)
			if err != nil {
				log.Fatalln(errors.WithStack(err))
			}
			patchFiles = make(map[string]string)
			patchPath := filepath.Join(rootPath, api.ManifestDirectory, PatchFolder)
			if _, err := os.Stat(patchPath); !os.IsNotExist(err) {
				err = filepath.Walk(patchPath, visitPatchFolder)
				if err != nil {
					log.Fatalln(errors.WithStack(err))
				}
			}
			err = filepath.Walk(filepath.Join(rootPath, api.ManifestDirectory, _VendorFolder), visitPatchAndDump)
			if err != nil {
				log.Fatalln(errors.WithStack(err))
			}
		},
	}

	cmd.Flags().StringVar(&src, "src", src, "Compile patch and source.")
	cmd.Flags().StringVar(&patch, "patch", patch, "Compile patch and source.")

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

	if fileInfo.Name() == api.DependencyFile {
		return nil
	}

	if strings.Contains(path, PatchFolder) {
		return nil
	}

	srcFilepath := path
	srcYamlByte, err := ioutil.ReadFile(srcFilepath)
	if err != nil {
		return errors.WithStack(err)
	}
	patchFileName, err := getPatchFileName(srcYamlByte)

	patchFilePath := patchFiles[patchFileName]
	if _, err := os.Stat(patchFilePath); err != nil {
		err = validator.ValidateBytes(srcYamlByte)
		if err != nil && !strings.Contains(path, PatchFolder) && strings.HasSuffix(path, ".jsonnet") {
			vm := jsonnet.MakeVM()
			j, err := vm.EvaluateSnippet(path, string(srcYamlByte))
			if err != nil {
				return errors.Wrap(err, "Error to evaluate jsonet")
			}
			yml, err := yaml.JSONToYAML([]byte(j))
			if err != nil {
				return errors.Wrap(err, "Error to evaluate jsonet: convert JSONToYAML")
			}
			srcYamlByte = yml
		}
		err = DumpCompiledFile(srcYamlByte, strings.Replace(path, _VendorFolder, CompileDirectory, 1))
		if err != nil {
			return errors.Wrap(err, "Error to evaluate jsonet: DumpCompiledFile")
		}
		return nil
	}

	patchByte, err := ioutil.ReadFile(patchFilePath)
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

func visitPatchFolder(path string, fileInfo os.FileInfo, ferr error) error {
	if ferr != nil {
		return ferr
	}
	if fileInfo.IsDir() {
		return nil
	}
	if !strings.Contains(path, PatchFolder) {
		return nil
	}
	patchFiles[fileInfo.Name()] = path
	return nil
}

func CompileWithpatchByPath(src, patch string) ([]byte, error) {
	srcYml, err := ioutil.ReadFile(src)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	patchYml, err := ioutil.ReadFile(patch)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	compiledYml, err := CompileWithPatch(srcYml, patchYml)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return compiledYml, nil
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

	match, err := checkGVKN(jsonSrc, jsonPatch)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !match {
		return nil, nil
	}

	var ro runtime.TypeMeta
	if err := yaml.Unmarshal(srcByte, &ro); err != nil {
		return nil, errors.WithStack(err)
	}
	kind := ro.GetObjectKind().GroupVersionKind()
	versionedObject, err := scheme.Scheme.New(kind)
	var compiled []byte
	switch {
	case runtime.IsNotRegisteredError(err):
		compiled, err = jsonpatch.MergePatch(jsonSrc, jsonPatch)
	case err != nil:
		return nil, errors.WithStack(err)
	default:
		compiled, err = strategicpatch.StrategicMergePatch(jsonSrc, jsonPatch, versionedObject)
	}
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
	if strings.Count(outlookPath, _VendorFolder) > 0 || strings.Count(outlookPath, CompileDirectory) > 1 || strings.Count(outlookPath, PatchFolder) > 0 {
		return nil
	}
	root := rootPath
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

	err = WriteCompiledFileToDest(outlookPath, annotateYaml)

	return nil
}

func WriteCompiledFileToDest(path string, compiledYaml []byte) error {
	_, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "Error to create outlook.")
	}

	err = ioutil.WriteFile(path, compiledYaml, 0755)
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

func convertJsonnetToYamlByFilepath(path string, srcYamlByte []byte) ([]byte, error) {
	vm := jsonnet.MakeVM()
	j, err := vm.EvaluateSnippet(path, string(srcYamlByte))
	if err != nil {
		return nil, errors.Wrap(err, "Error to evaluate jsonet")
	}
	yml, err := yaml.JSONToYAML([]byte(j))
	if err != nil {
		return nil, errors.Wrap(err, "Error to evaluate jsonet")
	}
	return yml, nil
}

// Decode decodes a list of objects in byte array format
func Decode(in []byte) ([]*unstructured.Unstructured, error) {
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(in), 1024)
	objs := []*unstructured.Unstructured{}

	var err error
	for {
		var out unstructured.Unstructured
		err = decoder.Decode(&out)
		if err != nil {
			break
		}
		objs = append(objs, &out)
	}
	if err != io.EOF {
		return nil, err
	}
	return objs, nil
}

func checkGVKN(srcJson, patchJson []byte) (bool, error) {
	src, err := Decode(srcJson)
	if err != nil {
		return false, errors.WithStack(err)
	}

	patch, err := Decode(patchJson)
	if err != nil {
		return false, errors.WithStack(err)
	}

	srcResource := &resource.Resource{
		Data: src[0],
	}
	patchResource := resource.Resource{
		Data: patch[0],
	}

	srcGvkn := srcResource.GVKN()
	patchGvkn := patchResource.GVKN()

	if srcGvkn == patchGvkn {
		return true, nil
	}
	return false, nil
}
