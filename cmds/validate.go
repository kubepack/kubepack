package cmds

import (
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"github.com/googleapis/gnostic/OpenAPIv2"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi/validation"
	"os"
	"io/ioutil"
	"path/filepath"
	"github.com/kubepack/kubepack/type"
	"net/http"
	"strings"
	"io"
	"sync"
	"gopkg.in/yaml.v2"
	"github.com/googleapis/gnostic/compiler"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi"
	"github.com/pkg/errors"
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
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	validator, err = GetOpenapiValidator(cmd)
	if err != nil {
		return err
	}

	outlookFolderpath := filepath.Join(path, CompileDirectory)
	_, err = os.Stat(outlookFolderpath)
	if os.IsNotExist(err) {
		return err
	}
	err = filepath.Walk(outlookFolderpath, visitOutlookFolder)
	if err != nil {
		return err
	}
	return nil
}

func NewFactory(cmd *cobra.Command) cmdutil.Factory {
	context := cmdutil.GetFlagString(cmd, "kube-context")
	config := configForContext(context)
	return cmdutil.NewFactory(config)
}

func configForContext(context string) clientcmd.ClientConfig {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}

	if context != "" {
		overrides.CurrentContext = context
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
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
		return "", fmt.Errorf("unable to get URL %q: %s", url, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unable to fetch file. URL: %q Status: %v", url, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read content of URL %q: %s", url, err.Error())
	}
	return strings.TrimSpace(string(body)), nil
}

func downloadOpenApiSwagger(url string, file *os.File) error {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return err
	}
	defer response.Body.Close()

	n, err := io.Copy(file, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return err
	}

	fmt.Println(n, "bytes downloaded.")
	return nil
}

func GetOpenapiValidator(cmd *cobra.Command) (*validation.SchemaValidation, error) {
	swaggerJsonPath, err := GetSwaggerJsonpath(cmd)
	if err != nil {
		return nil, err
	}
	apiSchema := ApiSchema{Path: swaggerJsonPath}
	doc, err := apiSchema.OpenApiSchema()
	if err != nil {
		return nil, err
	}
	resources, err := openapi.NewOpenAPIData(doc)
	return validation.NewSchemaValidation(resources), nil
}

func GetSwaggerJsonpath(cmd *cobra.Command) (string, error) {
	kubepackPath := filepath.Join(os.Getenv("HOME"), types.KubepackOpenapiPath)
	_, err := os.Stat(kubepackPath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(kubepackPath, 0755)
		if err != nil {
			return "", err
		}
	}

	var version string
	version, err = cmd.Flags().GetString("kube-version")
	if err != nil {
		return "", err
	}

	if version == "" {
		version, err = GetKubernetesVersion()
		if err != nil {
			return "", err
		}
	}

	oApiPath := filepath.Join(kubepackPath, version, "openapi-spec")
	_, err = os.Stat(oApiPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(oApiPath, 0755)
		if err != nil {
			return "", err
		}
	}

	swaggerpath := filepath.Join(oApiPath, "swagger.json")
	_, err = os.Stat(swaggerpath)
	if os.IsNotExist(err) {
		file, err := os.Create(swaggerpath)
		if err != nil {
			return "", err
		}
		defer file.Close()

		err = downloadOpenApiSwagger(fmt.Sprintf(OpenapiSpecDloadPath, version), file)
		if err != nil {
			return "", err
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
		return nil, err
	}
	spec, err := ioutil.ReadFile(f.Path)
	if err != nil {
		f.err = err
		return nil, err
	}
	var info yaml.MapSlice
	err = yaml.Unmarshal(spec, &info)
	if err != nil {
		f.err = err
		return nil, err
	}
	f.document, f.err = openapi_v2.NewDocument(info, compiler.NewContext("$root", nil))
	return f.document, f.err
}
