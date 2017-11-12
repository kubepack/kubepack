package copy

import (
	"fmt"
	"github.com/ghodss/yaml"
	typ "github.com/packsh/demo-dep/type"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type manifestStr struct {
	root string
}

func Copy(root string) error {

	manifest := filepath.Join(root, "manifest")
	fmt.Println("Hello world!!!", filepath.Join(root, "_vendor"))
	if _, err := os.Lstat(manifest); err != nil {
		err = os.Mkdir(filepath.Join(root, "manifest"), 0777)
		if err == nil {
			log.Println("Manifest successfully created...")
		} else {
			log.Println("Error occured....", err)
		}
	}
	str := &manifestStr{}
	str.root = root
	err := filepath.Walk(filepath.Join(root, "_vendor"), str.copyCallback)
	if err != nil {
		return err
	}
	return nil
}

func (a manifestStr) copyCallback(path string, info os.FileInfo, err error) error {
	man := filepath.Join(a.root, "manifest.yaml")
	byt, err := ioutil.ReadFile(man)
	manStruc := typ.ManifestDef{}
	err = yaml.Unmarshal(byt, &manStruc)
	for _, val := range manStruc.Dependencies {
		tmpPath := filepath.Join(val.Package, val.Folder)
		if strings.Contains(path, tmpPath) && val.Folder != "" {
			fmt.Println("Hello folder-----", tmpPath)
			fmt.Println("hello Path", path)
			fmt.Println("--------------------")
		}
	}
	if err != nil {
		log.Fatalln("Error Occuered-----", err)
	}

	imports := make([]string, len(manStruc.Dependencies))

	for key, value := range manStruc.Dependencies {
		imports[key] = value.Package
	}
	if info.IsDir() {
		// fmt.Println("Hello directory...", info.Name())
	} else {
		// fmt.Println("Hello file", info.Name())
	}
	return nil
}
