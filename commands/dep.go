package commands

import (
	"context"
	"flag"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/golang/dep/gps"
	"github.com/golang/dep/gps/pkgtree"
	"github.com/golang/glog"
	"github.com/google/go-jsonnet"
	api "github.com/kubepack/pack-server/apis/manifest/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	imports        []string
	packagePatches map[string]string
	forkRepo       []string
	finalFork      map[string]string
)

const PackTempDirectory = ".pack"

var depPatchFiles map[string][]string

type ApplyPatchToForkRepo struct {
	repo     string
	rootPath string
}

func NewDepCommand(plugin bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dep",
		Short: "Pulls dependent app manifests",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			validator, err = GetOpenapiValidator(cmd)
			if err != nil {
				log.Fatalln(err)
			}
			err = runDeps(cmd, plugin)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	flag.CommandLine.Parse([]string{})
	return cmd
}

func runDeps(cmd *cobra.Command, plugin bool) error {
	// Assume the current directory is correctly placed on a GOPATH, and that it's the
	// root of the project.
	packagePatches = make(map[string]string)
	depPatchFiles = make(map[string][]string)
	finalFork = make(map[string]string)
	logger := log.New(ioutil.Discard, "", 0)
	if glog.V(glog.Level(1)) {
		logger = log.New(os.Stdout, "", 0)
	}
	root, err := cmd.Flags().GetString("file")
	if err != nil {
		return errors.WithStack(err)
	}
	if !plugin && !filepath.IsAbs(root) {
		wd, err := os.Getwd()
		if err != nil {
			return errors.WithStack(err)
		}
		root = filepath.Join(wd, root)
	}
	if !filepath.IsAbs(root) {
		return errors.Errorf("Duh! we need an absolute path when used as a kubectl plugin. For more info, see here: https://github.com/kubernetes/kubectl/issues/346")
	}

	manifestPath := filepath.Join(root, api.DependencyFile)
	byt, err := ioutil.ReadFile(manifestPath)
	manStruc := api.DependencyList{}
	err = yaml.Unmarshal(byt, &manStruc)
	if err != nil {
		return errors.WithStack(err)
	}

	importroot := GetImportRoot(root)
	imports = make([]string, len(manStruc.Items))
	for key, value := range manStruc.Items {
		imports[key] = value.Package
	}
	manifestYaml := ManifestYaml{}
	manifestYaml.root = root
	pkgTree := map[string]pkgtree.PackageOrErr{
		"github.com/sdboyer/gps": {
			P: pkgtree.Package{
				Name:       "main",
				ImportPath: importroot,
				Imports:    imports,
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
		return errors.WithStack(err)
	}
	srcManagerConfig := gps.SourceManagerConfig{
		Cachedir:       filepath.Join(tempdir),
		Logger:         logger,
		DisableLocking: true,
	}
	log.Println("Tempdir: ", tempdir)
	sourcemgr, err := gps.NewSourceManager(srcManagerConfig)
	if err != nil {
		return errors.WithStack(err)
	}
	defer sourcemgr.Release()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Minute)
	defer cancel()
	// Prep and run the solver
	solver, err := gps.Prepare(params, sourcemgr)
	if err != nil {
		return errors.WithStack(err)
	}
	solution, err := solver.Solve(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	if err == nil {
		// If no failure, blow away the vendor dir and write a new one out,
		// stripping nested vendor directories as we go.
		vendorPath := filepath.Join(root, api.ManifestDirectory, _VendorFolder)
		err = os.RemoveAll(vendorPath)
		if err != nil {
			return errors.WithStack(err)
		}
		cascadingPruneOptions := gps.CascadingPruneOptions{}
		err = gps.WriteDepTree(vendorPath, solution, sourcemgr, cascadingPruneOptions, writeProgressLogger)
		if err != nil {
			return errors.WithStack(err)
		}

		err = filepath.Walk(vendorPath, findJsonnetFiles)
		if err != nil {
			return errors.WithStack(err)
		}

		// fork repository handling
		err = moveForkedDirectory(root)
		if err != nil {
			return errors.WithStack(err)
		}
		err = filepath.Walk(vendorPath, findPatchFolder)
		if err != nil {
			return errors.WithStack(err)
		}
		err = filepath.Walk(vendorPath, visitVendorAndApplyPatch)
		if err != nil {
			return errors.WithStack(err)
		}
		for _, val := range forkRepo {
			tmpDir := filepath.Join(os.Getenv("HOME"), PackTempDirectory, val)
			applypatch := &ApplyPatchToForkRepo{
				repo:     val,
				rootPath: root,
			}
			err = filepath.Walk(filepath.Join(tmpDir, api.ManifestDirectory), applypatch.applyPatchToFork)
			if err != nil {
				return errors.WithStack(err)
			}
			dstPath := filepath.Join(root, api.ManifestDirectory, _VendorFolder, val)
			finalFork[dstPath] = tmpDir
		}
		for key, val := range finalFork {
			err = os.RemoveAll(key)
			if err != nil {
				return errors.WithStack(err)
			}
			err = os.Rename(val, key)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}
	return nil
}

func findPatchFolder(path string, fileInfo os.FileInfo, ferr error) error {
	if ferr != nil {
		return errors.WithStack(ferr)
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

	// e.g:  _vendor/github.com/kubepack/kube-a/patch/github.com/kubepack/kube-a/nginx-b.deployment.apps.yaml
	// forkDir = github.com/kubepack/kube-a
	// patchFilePath = github.com/kubepack/kube-a/<name>.<kind>.<group>.yaml

	splitPatch := strings.Split(path, PatchFolder)
	patchFilePath := strings.TrimPrefix(splitPatch[1], "/")
	manifestsPath := strings.Join([]string{"/", "/"}, api.ManifestDirectory)
	if val, ok := packagePatches[patchFilePath]; ok {
		if val != strings.TrimSuffix(strings.TrimPrefix(forkDir, "/"), manifestsPath) {
			return nil
		}
	}
	pkg := filepath.Dir(patchFilePath)
	if _, ok := packagePatches[pkg]; ok {
		if !findImportInSlice(pkg, forkRepo) {
			forkRepo = append(forkRepo, pkg)
		}
	}
	depPatchFiles[pkg] = append(depPatchFiles[pkg], path)
	return nil
}

func (tmp *ApplyPatchToForkRepo) applyPatchToFork(path string, fileInfo os.FileInfo, ferr error) error {
	if ferr != nil {
		return ferr
	}
	if strings.Contains(path, PatchFolder) {
		return nil
	}
	if fileInfo.IsDir() {
		return nil
	}
	if fileInfo.Name() == api.DependencyFile {
		return nil
	}
	repoName := tmp.repo
	srcYaml, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.WithStack(err)
	}
	patchFileName, err := getPatchFileName(srcYaml)
	if err != nil {
		return errors.WithStack(err)
	}
	if _, ok := depPatchFiles[repoName]; !ok {
		return nil
	}
	for _, val := range depPatchFiles[repoName] {
		patchFile, err := os.Stat(val)
		if err != nil {
			return errors.WithStack(err)
		}
		if patchFileName == patchFile.Name() {
			patchYaml, err := ioutil.ReadFile(val)
			if err != nil {
				return errors.WithStack(err)
			}
			cmpldYaml, err := CompileWithPatch(srcYaml, patchYaml)
			if err != nil {
				return errors.WithStack(err)
			}
			err = WriteCompiledFileToDest(path, cmpldYaml)
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		}
	}
	return nil
}

func moveForkedDirectory(root string) error {
	packTmpDir := filepath.Join(os.Getenv("HOME"), ".pack")
	if _, err := os.Stat(packTmpDir); err == nil {
		err := os.RemoveAll(packTmpDir)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	err := os.MkdirAll(packTmpDir, 0755)
	if err != nil {
		return errors.WithStack(err)
	}
	for key := range packagePatches {
		forkPath := filepath.Join(root, api.ManifestDirectory, _VendorFolder, key, api.ManifestDirectory, _VendorFolder, key)
		fileInfo, err := os.Stat(forkPath)
		if err != nil {
			return errors.WithStack(err)
		}
		if !fileInfo.IsDir() {
			return errors.Errorf("Forked repository isn't a directory!!")
		}
		err = os.MkdirAll(filepath.Join(packTmpDir, filepath.Dir(key)), 0755)
		if err != nil {
			return errors.WithStack(err)
		}
		err = os.Rename(forkPath, filepath.Join(packTmpDir, key))
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func findJsonnetFiles(path string, fileInfo os.FileInfo, err error) error {
	if err != nil {
		return errors.WithStack(err)
	}
	if strings.Contains(path, PatchFolder) {
		return nil
	}
	if fileInfo.IsDir() {
		return nil
	}
	if strings.HasSuffix(path, "jsonnet.TEMPLATE") {
		return nil
	}
	if strings.HasSuffix(path, ".jsonnet") {
		srcYaml, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.WithStack(err)
		}
		err = validator.ValidateBytes(srcYaml)
		if err != nil {
			err = convertJsonnetfileToYamlfile(path)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	}

	return nil
}

func convertJsonnetfileToYamlfile(path string) error {
	vm := jsonnet.MakeVM()
	byt, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.WithStack(err)
	}
	j, err := vm.EvaluateSnippet(path, string(byt))
	if err != nil {
		return errors.WithStack(err)
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return errors.WithStack(err)
	}
	yml, err := convertJsonnetToYamlByFilepath(path, []byte(j))
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = f.Write([]byte(yml))
	if err != nil {
		return errors.WithStack(err)
	}

	err = os.Rename(path, path+".yaml")
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func findImportInSlice(r string, repos []string) bool {
	for _, val := range repos {
		if r == val {
			return true
		}
	}
	return false
}

func visitVendorAndApplyPatch(path string, fileInfo os.FileInfo, ferr error) error {
	if ferr != nil {
		return ferr
	}
	if strings.Contains(path, PatchFolder) {
		return nil
	}
	if fileInfo.IsDir() {
		return nil
	}
	if !strings.Contains(path, filepath.Join(api.ManifestDirectory, _VendorFolder)) {
		return nil
	}
	if strings.Count(path, filepath.Join(api.ManifestDirectory, _VendorFolder)) > 1 {
		return nil
	}
	if fileInfo.Name() == ".gitignore" || strings.HasSuffix(fileInfo.Name(), "jsonnet.TEMPLATE") {
		return nil
	}

	// e.g. path: /home/tigerworks/go/src/github.com/kubepack/pack/docs/_testdata/test-2/manifests/vendor/github.com/kubepack/kube-c/manifests/app/nginx-deployment.yaml
	// repoName := /github.com/kubepack/kube-c/manifests/app/nginx-deployment.yaml
	// repoName = /github.com/kubepack/kube-c/
	// pkg := github.com/kubepack/kube-c

	repoName := strings.Split(path, filepath.Join(api.ManifestDirectory, _VendorFolder))[1]
	repoName = strings.Split(repoName, api.ManifestDirectory)[0]
	pkg := strings.Trim(repoName, "/")
	if _, ok := depPatchFiles[pkg]; !ok {
		return nil
	}
	patches := depPatchFiles[pkg]
	for _, val := range patches {
		patchFile, err := os.Stat(val)
		if err != nil {
			return errors.WithStack(err)
		}
		patchName, err := getPatchFileNameByPath(path)
		if err != nil {
			return errors.WithStack(err)
		}
		if patchFile.Name() == patchName {
			mergedYml, err := CompileWithpatchByPath(path, val)
			err = WriteCompiledFileToDest(path, mergedYml)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}
	return nil
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
		return nil, nil, errors.Errorf("No directory exists at %s; cannot produce ProjectInfo", path)
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
		Name:    "kubepack",
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

	mpath := filepath.Join(a.root, api.DependencyFile)
	byt, err := ioutil.ReadFile(mpath)
	manStruc := api.DependencyList{}
	err = yaml.Unmarshal(byt, &manStruc)
	if err != nil {
		log.Fatalln("Error Occuered", err)
	}

	for _, value := range manStruc.Items {
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

	man := filepath.Join(a.root, api.DependencyFile)
	byt, err := ioutil.ReadFile(man)
	manStruc := api.DependencyList{}
	err = yaml.Unmarshal(byt, &manStruc)
	if err != nil {
		log.Fatalln(err)
	}

	for _, value := range manStruc.Items {
		properties := gps.ProjectProperties{}
		if value.Repo != "" {
			properties.Source = value.Repo
		} else if value.Fork != "" {
			if _, ok := packagePatches[value.Package]; ok {
				log.Fatal(errors.Errorf("%s defined in multiple packages.", value.Package))
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

func (a ManifestYaml) TestDependencyConstraints() gps.ProjectConstraints {
	return nil
}

type InternalManifest struct {
	root string
}

func (a InternalManifest) DependencyConstraints() gps.ProjectConstraints {
	projectConstraints := make(gps.ProjectConstraints)
	man := filepath.Join(a.root, api.DependencyFile)
	byt, err := ioutil.ReadFile(man)
	manStruc := api.DependencyList{}
	err = yaml.Unmarshal(byt, &manStruc)
	if err != nil {
		log.Fatalln("Error Occuered-----", err)
	}

	for _, value := range manStruc.Items {
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
	mpath := filepath.Join(root, api.DependencyFile)
	if _, err := os.Lstat(mpath); err != nil {
		return nil, nil, errors.Wrap(err, "Unable to read manifest.yaml")
	}
	man := &InternalManifest{}
	man.root = root
	lck := &InternalLock{}
	lck.root = root
	return man, lck, nil
}

func GetImportRoot(root string) string {
	srcprefix := filepath.Join(build.Default.GOPATH, "src") + string(filepath.Separator)
	importroot := filepath.ToSlash(strings.TrimPrefix(root, srcprefix))
	return importroot
}

func writeProgressLogger(progress gps.WriteProgress) {
	glog.Infof("repo(%d/%d):  %s\n", progress.Count, progress.Total, progress.LP)
}
