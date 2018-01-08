package cmds

import (
	"context"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/appscode/go/crypto/rand"
	"github.com/ghodss/yaml"
	"github.com/golang/dep/gps"
	"github.com/golang/dep/gps/pkgtree"
	"github.com/golang/glog"
	typ "github.com/kubepack/kubepack/type"
	"github.com/spf13/cobra"
)

var (
	imports        []string
	packagePatches map[string]string
	forkRepo       []string
)

func NewDepCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dep",
		Short: "Pulls dependent app manifests",
		Run: func(cmd *cobra.Command, args []string) {
			err := runDeps()
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	return cmd
}

func runDeps() error {
	// Assume the current directory is correctly placed on a GOPATH, and that it's the
	// root of the project.
	packagePatches = make(map[string]string)
	logger := log.New(ioutil.Discard, "", 0)
	if glog.V(glog.Level(1)) {
		logger = log.New(os.Stdout, "", 0)
	}
	root, _ := os.Getwd()
	manifestPath := filepath.Join(root, typ.ManifestFile)
	byt, err := ioutil.ReadFile(manifestPath)
	manStruc := typ.ManifestDefinition{}
	err = yaml.Unmarshal(byt, &manStruc)
	if err != nil {
		return err
	}

	imports = make([]string, len(manStruc.Dependencies))

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
		TraceLogger:     logger,
		ProjectAnalyzer: NaiveAnalyzer{},
		Manifest:        manifestYaml,
		RootPackageTree: pkgtree.PackageTree{
			ImportRoot: importroot,
			Packages:   pkgTree,
		},
	}
	// Set up a SourceManager. This manages interaction with sources (repositories).
	tempdir, err := ioutil.TempDir("", "gps-repocache")
	if err != nil {
		return err
	}
	srcManagerConfig := gps.SourceManagerConfig{
		Cachedir:       filepath.Join(tempdir),
		Logger:         logger,
		DisableLocking: true,
	}
	log.Println("Tempdir: ", tempdir)
	sourcemgr, _ := gps.NewSourceManager(srcManagerConfig)
	defer sourcemgr.Release()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Minute)
	defer cancel()
	// Prep and run the solver
	solver, err := gps.Prepare(params, sourcemgr)
	if err != nil {
		return err
	}
	solution, err := solver.Solve(ctx)
	if err != nil {
		return err
	}
	if err == nil {
		// If no failure, blow away the vendor dir and write a new one out,
		// stripping nested vendor directories as we go.
		err = os.RemoveAll(filepath.Join(root, _VendorFolder))
		if err != nil {
			return err
		}
		err = gps.WriteDepTree(filepath.Join(root, _VendorFolder), solution, sourcemgr, true, logger)
		if err != nil {
			return err
		}

		err = filepath.Walk(filepath.Join(root, _VendorFolder), findPatchFolder)
		if err != nil {
			return err
		}

		for _, val := range forkRepo {
			srcPath := filepath.Join(root, _VendorFolder, val, _VendorFolder, val)
			if _, err = os.Stat(srcPath); err != nil {
				return nil
			}
			tmpDir := filepath.Join(root, _VendorFolder, rand.WithUniqSuffix("hello"))
			err = os.Rename(srcPath, tmpDir)
			if err != nil {
				return err
			}

			dstPath := filepath.Join(root, _VendorFolder, val)

			err = os.RemoveAll(dstPath)
			if err != nil {
				return err
			}

			err = os.Rename(tmpDir, dstPath)
		}
	}
	return nil
}

func findPatchFolder(path string, fileInfo os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !strings.Contains(path, PatchFolder) {
		return nil
	}
	if fileInfo.IsDir() {
		return nil
	}

	if strings.Index(path, _VendorFolder) != strings.LastIndex(path, _VendorFolder) {
		return nil
	}

	splitVendor := strings.Split(path, _VendorFolder)
	forkDir := strings.TrimPrefix(strings.Split(splitVendor[1], PatchFolder)[0], "/")
	splitPatch := strings.Split(path, PatchFolder)
	patchFilePath := strings.TrimPrefix(splitPatch[1], "/")
	srcDir := filepath.Join(strings.Split(path, _VendorFolder)[0], _VendorFolder, patchFilePath)
	if val, ok := packagePatches[patchFilePath]; ok {
		if val != strings.TrimSuffix(strings.TrimPrefix(forkDir, "/"), "/") {
			return nil
		}
	}
	pkg := strings.TrimSuffix(forkDir, "/")
	if _, ok := packagePatches[pkg]; ok {
		src := strings.Replace(path, PatchFolder, _VendorFolder, 1)
		srcDir = src
		if !findImportInSlice(pkg, forkRepo) {
			forkRepo = append(forkRepo, pkg)
		}
	}

	if _, err := os.Stat(srcDir); err == nil {
		srcYaml, err := ioutil.ReadFile(srcDir)
		if err != nil {
			return err
		}

		patchYaml, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		mergedYaml, err := CompileWithPatch(srcYaml, patchYaml)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(srcDir, mergedYaml, 0755)
		if err != nil {
			return err
		}
	}

	return err
}

func findImportInSlice(r string, repos []string) bool {
	for _, val := range repos {
		if r == val {
			return true
		}
	}
	return false
}

type NaiveAnalyzer struct {
}

// DeriveManifestAndLock is called when the solver needs manifest/lock data
// for a particular dependency project (identified by the gps.ProjectRoot
// parameter) at a particular version. That version will be checked out in a
// directory rooted at path.
func (a NaiveAnalyzer) DeriveManifestAndLock(path string, n gps.ProjectRoot) (gps.Manifest, gps.Lock, error) {
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
		Name:    "kubernetes-dependency-mngr",
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
	ovrr := gps.ProjectConstraints{}

	mpath := filepath.Join(a.root, typ.ManifestFile)
	byt, err := ioutil.ReadFile(mpath)
	manStruc := typ.ManifestDefinition{}
	err = yaml.Unmarshal(byt, &manStruc)
	if err != nil {
		log.Fatalln("Error Occuered-----", err)
	}

	for _, value := range manStruc.Dependencies {
		properties := gps.ProjectProperties{}
		if value.Repo != "" {
			properties.Source = value.Repo
		} else if value.Fork != "" {
			properties.Source = value.Fork
		} else {
			properties.Source = value.Package
		}
		if value.Branch != "" {
			properties.Constraint = gps.NewBranch(value.Branch)
		} else if value.Version != "" {
			properties.Constraint = gps.Revision(value.Version)
		}
		ovrr[gps.ProjectRoot(value.Package)] = properties
	}
	return ovrr
}

func (a ManifestYaml) DependencyConstraints() gps.ProjectConstraints {
	projectConstraints := make(gps.ProjectConstraints)

	man := filepath.Join(a.root, typ.ManifestFile)
	byt, err := ioutil.ReadFile(man)
	manStruc := typ.ManifestDefinition{}
	err = yaml.Unmarshal(byt, &manStruc)
	if err != nil {
		log.Fatalln(err)
	}

	for _, value := range manStruc.Dependencies {
		properties := gps.ProjectProperties{}
		if value.Repo != "" {
			properties.Source = value.Repo
		} else if value.Fork != "" {
			if _, ok := packagePatches[value.Package]; ok {
				log.Fatal(fmt.Errorf("%s defined in multiple packages.", value.Package))
			}
			packagePatches[value.Package] = value.Fork
			properties.Source = value.Fork
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

func mapPatches(repo string, patches []string) error {
	for _, val := range patches {
		if _, ok := packagePatches[val]; ok {
			return fmt.Errorf("%s defined in multiple packages.", val)
		}
		packagePatches[val] = repo
	}
	return nil
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
	return nil
}

func (a InternalLock) InputsDigest() []byte {
	return nil
}

func (a NaiveAnalyzer) lookForManifest(root string) (gps.Manifest, gps.Lock, error) {
	mpath := filepath.Join(root, typ.ManifestFile)
	if _, err := os.Lstat(mpath); err != nil {
		return nil, nil, err
	}
	man := &InternalManifest{}
	man.root = root
	lck := &InternalLock{}
	lck.root = root
	return man, lck, nil
}
