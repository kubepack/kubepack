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
	// "github.com/packsh/demo-dep/dep/copy"
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
	man := filepath.Join(root, "manifest.yaml")
	byt, err := ioutil.ReadFile(man)
	manStruc := typ.ManifestDefinition{}
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

	manifestYaml := ManifestYaml{}
	manifestYaml.root = root
	pkgTree := map[string]pkgtree.PackageOrErr{
		"github.com/sdboyer/gps": {
			P: pkgtree.Package{
				// Name:       "github.com/a8uhnf/go_stack",
				// ImportPath: "github.com/packsh/demo-dep",
				Imports: imports,
			},
		},
	}
	params := gps.SolveParameters{
		RootDir:         root,
		TraceLogger:     log.New(os.Stdout, "", 0),
		ProjectAnalyzer: NaiveAnalyzer{},
		Manifest:        manifestYaml,
		RootPackageTree: pkgtree.PackageTree{
			ImportRoot: importroot,
			Packages:   pkgTree,
		},
	}
	// Set up a SourceManager. This manages interaction with sources (repositories).
	tempdir, _ := ioutil.TempDir("", "gps-repocache")
	srcManagerConfig := gps.SourceManagerConfig{
		Cachedir:       filepath.Join(tempdir),
		Logger:         log.New(os.Stdout, "", 0),
		DisableLocking: true,
	}
	log.Println("hello tempdir", tempdir)
	sourcemgr, _ := gps.NewSourceManager(srcManagerConfig)
	defer sourcemgr.Release()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Minute)
	defer cancel()
	// Prep and run the solver
	solver, err := gps.Prepare(params, sourcemgr)
	if err != nil {
		fmt.Println("Prepare error occurred..", err)
		return
	}
	solution, err := solver.Solve(ctx)
	if err != nil {
		fmt.Println("Solve error occurred..", err)
		return
	}
	if err == nil {
		// If no failure, blow away the vendor dir and write a new one out,
		// stripping nested vendor directories as we go.
		os.RemoveAll(filepath.Join(root, "_vendor"))
		gps.WriteDepTree(filepath.Join(root, "_vendor"), solution, sourcemgr, true, log.New(os.Stdout, "Hello:-----", 4))
	}
}

type NaiveAnalyzer struct {
	// lookForManifest(root string) (gps.)
}

// DeriveManifestAndLock is called when the solver needs manifest/lock data
// for a particular dependency project (identified by the gps.ProjectRoot
// parameter) at a particular version. That version will be checked out in a
// directory rooted at path.
func (a NaiveAnalyzer) DeriveManifestAndLock(path string, n gps.ProjectRoot) (gps.Manifest, gps.Lock, error) {
	fmt.Println("hello world naive analyzer.........", path, n)
	// man := filepath.Join(filepath.Join(path, "manifest.yaml"))
	// return nil, nil, nil
	// this check should be unnecessary, but keeping it for now as a canary
	if _, err := os.Lstat(path); err != nil {
		return nil, nil, fmt.Errorf("No directory exists at %s; cannot produce ProjectInfo", path)
	}

	m, l, err := a.lookForManifest(path)
	if err == nil {
		// TODO verify project name is same as what SourceManager passed in?
		return m, l, nil
	} else {
		return nil, nil, err
	}
}

// Reports the name and version of the analyzer. This is used internally as part
// of gps' hashing memoization scheme.
func (a NaiveAnalyzer) Info() gps.ProjectAnalyzerInfo {
	return gps.ProjectAnalyzerInfo{
		Name:    "example-analyzer",
		Version: 1,
	}
}

type ManifestYaml struct {
	root string
}

func (a ManifestYaml) IgnoredPackages() *pkgtree.IgnoredRuleset {
	return nil
}

func (a ManifestYaml) RequiredPackages() map[string]bool {
	return nil
}

func (a ManifestYaml) Overrides() gps.ProjectConstraints {
	return nil
}

func (a ManifestYaml) DependencyConstraints() gps.ProjectConstraints {
	projectConstraints := make(gps.ProjectConstraints)

	man := filepath.Join(a.root, "manifest.yaml")
	byt, err := ioutil.ReadFile(man)
	manStruc := typ.ManifestDefinition{}
	err = yaml.Unmarshal(byt, &manStruc)
	if err != nil {
		log.Fatalln("Error Occuered-----", err)
	}

	for _, value := range manStruc.Dependencies {
		// fmt.Println("Hello key", key)
		// fmt.Println("Hello value package", value.Package)
		properties := gps.ProjectProperties{}
		if value.Repo != "" {
			properties.Source = value.Repo
		} else {
			properties.Source = value.Package
		}
		if value.Branch != "" {
			fmt.Println("Hello branch -----", value.Branch)
			properties.Constraint = gps.NewBranch(value.Branch)
		} else if value.Version != "" {
			properties.Constraint = gps.Revision(value.Version)
		}
		projectConstraints[gps.ProjectRoot(value.Package)] = properties
	}
	return projectConstraints
}

func (a ManifestYaml) TestDependencyConstraints() gps.ProjectConstraints {
	return nil
}

type InternalManifest struct {
	root string
}

func (a InternalManifest) DependencyConstraints() gps.ProjectConstraints {
	projectConstraints := make(gps.ProjectConstraints)

	man := filepath.Join(a.root, typ.ManifestFile)
	byt, err := ioutil.ReadFile(man)
	manStruc := typ.ManifestDefinition{}
	err = yaml.Unmarshal(byt, &manStruc)
	if err != nil {
		log.Fatalln("Error Occuered-----", err)
	}

	for _, value := range manStruc.Dependencies {
		properties := gps.ProjectProperties{}
		if value.Repo != "" {
			properties.Source = value.Repo
		} else {
			properties.Source = value.Package
		}
		if value.Branch != "" {
			properties.Constraint = gps.NewBranch(value.Branch)
		} else if value.Version != "" {
			properties.Constraint = gps.Revision(value.Version)
		}
		projectConstraints[gps.ProjectRoot(value.Package)] = properties
	}
	return projectConstraints
}

type InternalLock struct {
	root string
}

func (a InternalLock) Projects() []gps.LockedProject {
	man := filepath.Join(a.root, typ.ManifestFile)
	byt, err := ioutil.ReadFile(man)
	manStruc := typ.ManifestDefinition{}
	err = yaml.Unmarshal(byt, &manStruc)
	if err != nil {
		panic(err)
	}
	lockedProjs := make([]gps.LockedProject, len(manStruc.Dependencies))

	for key, value := range manStruc.Dependencies {
		// r := gps.Revision()
		// properties := gps.ProjectProperties{}
		ident := gps.ProjectIdentifier{
			ProjectRoot: gps.ProjectRoot(value.Package),
		}
		var v gps.Version
		if value.Repo != "" {
			ident.Source = value.Repo
		} else {
			ident.Source = value.Package
		}

		if value.Branch != "" {
			v = gps.NewBranch(value.Branch)
		} else if value.Version != "" {
			// properties.Constraint = gps.Revision(value.Version)
			v = gps.NewVersion(value.Version)
		}
		lockedProjs[key] = &gps.LockedProject{
			ident,
			v,
			"",
		}
	}
	return lockedProjs
}

func (a InternalLock) InputsDigest() []byte {
	return nil
}

func (a NaiveAnalyzer) lookForManifest(root string) (gps.Manifest, gps.Lock, error) {
	man := &InternalManifest{}
	man.root = root
	lck := &InternalLock{}
	lck.root = root
	return man, lck, nil
}
