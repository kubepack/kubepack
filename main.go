package main

import (
	"log"
	"os"
	"fmt"
	"path/filepath"
	"go/build"
	"strings"
	"github.com/sdboyer/gps/pkgtree"
)

func main()  {
	log.Println("Starting demo dep....")
	root, _ := os.Getwd()
	srcprefix := filepath.Join(build.Default.GOPATH, "src") + string(filepath.Separator)
	importroot := filepath.ToSlash(strings.TrimPrefix(root, srcprefix))
	fmt.Println("Get the root:-", root)
	fmt.Println("Get the importroot:-", importroot)
	pkgList, err := pkgtree.ListPackages(root, importroot)
	if err != nil {
		log.Fatalln("Error getting pkglist.....", err)
	}
	log.Println("Hello packageList", pkgList)
}
