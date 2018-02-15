package cmds

import (
	"testing"
	"fmt"
	"github.com/google/go-jsonnet"
	"io/ioutil"
	yaml1 "gopkg.in/yaml.v2"
	"github.com/ghodss/yaml"
)

func TestEvaluateJsonnetSnippet(t *testing.T)  {

	vm := jsonnet.MakeVM()

	jsonnetpath := "/home/tigerworks/go/src/github.com/databricks/jsonnet-style-guide/examples/databricks/simple/foocorp-manager.jsonnet"

	/// fmt.Println(jsonnetpath)
	byt, err := ioutil.ReadFile(jsonnetpath)
	if err != nil {
		panic(err)
	}


	output, err := vm.EvaluateSnippet(jsonnetpath, string(byt))
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
	fmt.Println("*&*&*&*&*&*&")
	// fmt.Println(output)
	// return

	/*out, err := yaml1.Marshal(output)
	if err != nil {
		panic(err)
	}*/
	y, err := yaml.JSONToYAML([]byte(output))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(y))
	// yaml1.Marshaler()
}
