package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"context"
	"time"

	"github.com/ghodss/yaml"
	"github.com/golang/dep/gps"
	"github.com/golang/dep/gps/pkgtree"
	"github.com/packsh/demo-dep/dep/copy"
	typ "github.com/packsh/demo-dep/type"
)

// This is probably the simplest possible implementation of gps. It does the
// substantive work that `go get` does, except:
//  1. It drops the resulting tree into vendor instead of GOPATH
//  2. It prefers semver tags (if available) over branches
//  3. It removes any vendor directories nested within dependencies
//
//  This will compile and work...and then blow away any vendor directory present
//  in the cwd. Be careful!

func main() {
	// Assume the current directory is correctly placed on a GOPATH, and that it's the
	// root of the project.
	root, _ := os.Getwd()
	err := copy.Copy(root)
	fmt.Println("Hello error", err)
	return
	man := filepath.Join(root, "manifest.yaml")
	byt, err := ioutil.ReadFile(man)
	manStruc := typ.ManifestDef{}
	err = yaml.Unmarshal(byt, &manStruc)
	if err != nil {
		log.Fatalln("Error Occuered-----", err)
	}

	imports := make([]string, len(manStruc.Dependencies))

	for key, value := range manStruc.Dependencies {
		imports[key] = value.Package
	}

	srcprefix := filepath.Join(build.Default.GOPATH, "src") + string(filepath.Separator)
	importroot := filepath.ToSlash(strings.TrimPrefix(root, srcprefix))

	params := gps.SolveParameters{
		RootDir: root,
		// Trace:           false,
		TraceLogger:     log.New(os.Stdout, "", 0),
		ProjectAnalyzer: NaiveAnalyzer{},
		Manifest:        ManifestYaml{},
		RootPackageTree: pkgtree.PackageTree{
			ImportRoot: importroot,
			Packages: map[string]pkgtree.PackageOrErr{
				"github.com/sdboyer/gps": pkgtree.PackageOrErr{
					P: pkgtree.Package{
						// Name:       "github.com/a8uhnf/go_stack",
						// ImportPath: "github.com/packsh/demo-dep",
						Imports: imports,
					},
				},
			},
		},
	}

	// Set up a SourceManager. This manages interaction with sources (repositories).
	tempdir, _ := ioutil.TempDir("", "gps-repocache")
	srcManagerConfig := gps.SourceManagerConfig{
		Cachedir:       filepath.Join(tempdir),
		Logger:         log.New(os.Stdout, "", 0),
		DisableLocking: true,
	}
	sourcemgr, _ := gps.NewSourceManager(srcManagerConfig)
	defer sourcemgr.Release()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Minute)
	defer cancel()
	// Prep and run the solver
	fmt.Println("Got it never")
	solver, err := gps.Prepare(params, sourcemgr)
	fmt.Println("Got it", err)
	solution, err := solver.Solve(ctx)
	fmt.Println("Hello Error", err)
	if err == nil {
		// If no failure, blow away the vendor dir and write a new one out,
		// stripping nested vendor directories as we go.
		os.RemoveAll(filepath.Join(root, "_vendor"))
		gps.WriteDepTree(filepath.Join(root, "_vendor"), solution, sourcemgr, true, log.New(os.Stdout, "Hello", 0))
		err = copy.Copy(root)
		fmt.Println("Hello error", err)
	}
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
func (a NaiveAnalyzer) Info() gps.ProjectAnalyzerInfo {
	return gps.ProjectAnalyzerInfo{
		Name:    "example-analyzer",
		Version: 1,
	}
}

type ManifestYaml struct{}

func (a ManifestYaml) IgnoredPackages() *pkgtree.IgnoredRuleset {
	return nil
}

func (a ManifestYaml) RequiredPackages() map[string]bool {
	return nil
}

func (a ManifestYaml) Overrides() gps.ProjectConstraints {
	// return nil
	return gps.ProjectConstraints{
		"github.com/Masterminds/semver": gps.ProjectProperties{
			Source:     "github.com/Masterminds/semver",
			Constraint: gps.NewBranch("2.x"),
		},
	}
}

func (a ManifestYaml) DependencyConstraints() gps.ProjectConstraints {
	return gps.ProjectConstraints{
		"github.com/appscode/go": gps.ProjectProperties{
			Source:     "github.com/appscode/go",
			Constraint: gps.NewBranch("master"),
		},
		"github.com/Masterminds/semver": gps.ProjectProperties{
			Source:     "github.com/Masterminds/semver",
			Constraint: gps.NewBranch("2.x"),
		},
	}
}

func (a ManifestYaml) TestDependencyConstraints() gps.ProjectConstraints {
	return nil
}
