package main

import (
	"log"
	"os"
	"fmt"
	"path/filepath"
	"go/build"
	"strings"
	"github.com/sdboyer/gps/pkgtree"
	"github.com/gps/gps"
)

func main()  {
	log.Println("Starting demo dep....")
	root, _ := os.Getwd()
	srcprefix := filepath.Join(build.Default.GOPATH, "src") + string(filepath.Separator)
	importroot := filepath.ToSlash(strings.TrimPrefix(root, srcprefix))
	fmt.Println("Get the root:-", root)
	fmt.Println("Get the importroot:-", importroot)

	params := gps.SolveParameters{
		RootDir: root,
		Trace: true,
		TraceLogger: log.New(os.Stdout, "", 0),
		ProjectAnalyzer: NaiveAnalyzer{},
	}
	pkgList, err := pkgtree.ListPackages(root, importroot)
	if err != nil {
		log.Fatalln("Error getting pkglist.....", err)
	}
	params.RootPackageTree = pkgList
	log.Println("Hello packageList", pkgList)
}

type NaiveAnalyzer struct{}
// DeriveManifestAndLock is called when the solver needs manifest/lock data
// for a particular dependency project (identified by the gps.ProjectRoot
// parameter) at a particular version. That version will be checked out in a
// directory rooted at path.
func (a NaiveAnalyzer) DeriveManifestAndLock(path string, n gps.ProjectRoot) (gps.Manifest, gps.Lock, error) {
	return nil, nil, nil
}

// Reports the name and version of the analyzer. This is used internally as part
// of gps' hashing memoization scheme.
func (a NaiveAnalyzer) Info() (name string, version int) {
	return "example-analyzer", 1
}
