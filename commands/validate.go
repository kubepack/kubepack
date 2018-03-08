package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/googleapis/gnostic/OpenAPIv2"
	"github.com/googleapis/gnostic/compiler"
	api "github.com/kubepack/pack-server/apis/manifest/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi/validation"
)

const OpenapiSpecDloadPath = "https://raw.githubusercontent.com/kubernetes/kubernetes/%s/api/openapi-spec/swagger.json"

var validator *validation.SchemaValidation

func NewValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate _outlook folder",
		Run: func(cmd *cobra.Command, args []string) {
			err := validateOutlook(cmd)
			if err != nil {
				panic(err)
			}
		},
	}
	return cmd
}

func validateOutlook(cmd *cobra.Command) error {
	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return errors.WithStack(err)
	}
	if !filepath.IsAbs(path) {
		return errors.Errorf("Need to provide Absolute path. Here is the issue: https://github.com/kubernetes/kubectl/issues/346")
	}
	validator, err = GetOpenapiValidator(cmd)
	if err != nil {
		return errors.WithStack(err)
	}

	outlookFolderpath := filepath.Join(path, CompileDirectory)
	_, err = os.Stat(outlookFolderpath)
	if os.IsNotExist(err) {
		return errors.WithStack(err)
	}
	err = filepath.Walk(outlookFolderpath, visitOutlookFolder)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func visitOutlookFolder(path string, fileInfo os.FileInfo, ferr error) error {
	if ferr != nil {
		return errors.WithStack(ferr)
	}
	if fileInfo.IsDir() {
		return nil
	}

	srcBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.WithStack(ferr)
	}

	err = validator.ValidateBytes(srcBytes)
	if err != nil {
		return errors.WithStack(ferr)
	}
	fmt.Printf("%s is a valid yaml\n", path)
	return nil
}

func GetKubernetesVersion() (string, error) {
	url := "https://dl.k8s.io/release/stable.txt"
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.Wrapf(err, "unable to get URL %q: %s", url, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("unable to fetch file. URL: %q Status: %v", url, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read content of URL %q: %s", url, err.Error())
	}
	return strings.TrimSpace(string(body)), nil
}

func downloadOpenApiSwagger(url string, file *os.File) error {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return errors.WithStack(err)
	}
	defer response.Body.Close()

	n, err := io.Copy(file, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return errors.WithStack(err)
	}

	fmt.Println(n, "bytes downloaded.")
	return nil
}

func GetOpenapiValidator(cmd *cobra.Command) (*validation.SchemaValidation, error) {
	swaggerJsonPath, err := GetSwaggerJsonpath(cmd)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	apiSchema := ApiSchema{Path: swaggerJsonPath}
	doc, err := apiSchema.OpenApiSchema()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resources, err := openapi.NewOpenAPIData(doc)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return validation.NewSchemaValidation(resources), nil
}

func GetSwaggerJsonpath(cmd *cobra.Command) (string, error) {
	kubepackPath := filepath.Join(os.Getenv("HOME"), api.KubepackOpenapiPath)
	_, err := os.Stat(kubepackPath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(kubepackPath, 0755)
		if err != nil {
			return "", errors.WithStack(err)
		}
	}

	var version string
	version, err = cmd.Flags().GetString("kube-version")
	if err != nil {
		return "", errors.WithStack(err)
	}

	if version == "" {
		version, err = GetKubernetesVersion()
		if err != nil {
			return "", errors.WithStack(err)
		}
	}

	oApiPath := filepath.Join(kubepackPath, version, "openapi-spec")
	_, err = os.Stat(oApiPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(oApiPath, 0755)
		if err != nil {
			return "", errors.WithStack(err)
		}
	}

	swaggerpath := filepath.Join(oApiPath, "swagger.json")
	_, err = os.Stat(swaggerpath)
	if os.IsNotExist(err) {
		file, err := os.Create(swaggerpath)
		if err != nil {
			return "", errors.WithStack(err)
		}
		defer file.Close()

		err = downloadOpenApiSwagger(fmt.Sprintf(OpenapiSpecDloadPath, version), file)
		if err != nil {
			return "", errors.WithStack(err)
		}
	}

	return swaggerpath, nil
}

type ApiSchema struct {
	Path     string
	once     sync.Once
	document *openapi_v2.Document
	err      error
}

func (f *ApiSchema) OpenApiSchema() (*openapi_v2.Document, error) {
	_, err := os.Stat(f.Path)
	if err != nil {
		f.err = err
		return nil, errors.WithStack(err)
	}
	spec, err := ioutil.ReadFile(f.Path)
	if err != nil {
		f.err = err
		return nil, errors.WithStack(err)
	}
	var info yaml.MapSlice
	err = yaml.Unmarshal(spec, &info)
	if err != nil {
		f.err = err
		return nil, errors.WithStack(err)
	}
	f.document, f.err = openapi_v2.NewDocument(info, compiler.NewContext("$root", nil))
	return f.document, f.err
}
