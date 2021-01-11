/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package action

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"kubepack.dev/cli/pkg/apply"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/kube"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/postrender"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/releaseutil"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/resource"
	"sigs.k8s.io/yaml"
)

// Apply performs an applyation operation.
type Apply struct {
	ApplyOptions *apply.ApplyOptions
	cfg          *Configuration

	ChartPathOptions

	ClientOnly               bool
	CreateNamespace          bool
	DryRun                   bool
	DisableHooks             bool
	Replace                  bool
	Wait                     bool
	Devel                    bool
	DependencyUpdate         bool
	Timeout                  time.Duration
	Namespace                string
	ReleaseName              string
	GenerateName             bool
	NameTemplate             string
	Description              string
	OutputDir                string
	Atomic                   bool
	SkipCRDs                 bool
	SubNotes                 bool
	DisableOpenAPIValidation bool
	IncludeCRDs              bool
	// APIVersions allows a manual set of supported API Versions to be passed
	// (for things like templating). These are ignored if ClientOnly is false
	APIVersions chartutil.VersionSet
	// Used by helm template to render charts with .Release.IsUpgrade. Ignored if Dry-Run is false
	IsUpgrade bool
	// Used by helm template to add the release as part of OutputDir path
	// OutputDir/<ReleaseName>
	UseReleaseName bool
	PostRenderer   postrender.PostRenderer
}

// NewApply creates a new Apply object with the given configuration.
func NewApply(cfg *Configuration) *Apply {
	return &Apply{
		cfg: cfg,
	}
}

func (i *Apply) applyCRDs(crds []chart.CRD) error {
	// We do these one file at a time in the order they were read.
	totalItems := []*resource.Info{}
	for _, obj := range crds {
		// Read in the resources
		res, err := i.cfg.KubeClient.Build(bytes.NewBuffer(obj.File.Data), false)
		if err != nil {
			return errors.Wrapf(err, "failed to apply CRD %s", obj.Name)
		}

		// Send them to Kube
		errs := []error{}
		for _, info := range res {
			if err := i.ApplyOptions.ApplyOneObject(info); err != nil {
				errs = append(errs, err)
			}
		}
		// If any errors occurred during apply, then return error (or
		// aggregate of errors).
		if len(errs) == 1 {
			return errs[0]
		}
		if len(errs) > 1 {
			return utilerrors.NewAggregate(errs)
		}

		//if _, err := i.cfg.KubeClient.Create(res); err != nil {
		//	// If the error is CRD already exists, continue.
		//	if apierrors.IsAlreadyExists(err) {
		//		crdName := res[0].Name
		//		i.cfg.Log("CRD %s is already present. Skipping.", crdName)
		//		continue
		//	}
		//	return errors.Wrapf(err, "failed to apply CRD %s", obj.Name)
		//}
		totalItems = append(totalItems, res...)
	}
	if len(totalItems) > 0 {
		// Invalidate the local cache, since it will not have the new CRDs
		// present.
		discoveryClient, err := i.cfg.RESTClientGetter.ToDiscoveryClient()
		if err != nil {
			return err
		}
		i.cfg.Log("Clearing discovery cache")
		discoveryClient.Invalidate()
		// Give time for the CRD to be recognized.

		if err := i.cfg.KubeClient.Wait(totalItems, 60*time.Second); err != nil {
			return err
		}

		// Make sure to force a rebuild of the cache.
		discoveryClient.ServerGroups()
	}
	return nil
}

// Run executes the applyation
//
// If DryRun is set to true, this will prepare the release, but not apply it
func (i *Apply) Run(chrt *chart.Chart, vals map[string]interface{}) (*release.Release, error) {
	// Check reachability of cluster unless in client-only mode (e.g. `helm template` without `--validate`)
	if !i.ClientOnly {
		if err := i.cfg.KubeClient.IsReachable(); err != nil {
			return nil, err
		}
	}

	// TAMAL: Avoid this check for upgrade flow
	//if err := i.availableName(); err != nil {
	//	return nil, err
	//}

	if err := validateReleaseName(i.ReleaseName); err != nil {
		return nil, errors.Errorf("release name is invalid: %s", i.ReleaseName)
	}
	i.cfg.Log("preparing upgrade for %s", i.ReleaseName)

	// Pre-apply anything in the crd/ directory. We do this before Helm
	// contacts the upstream server and builds the capabilities object.
	if crds := chrt.CRDObjects(); !i.ClientOnly && !i.SkipCRDs && len(crds) > 0 {
		// On dry run, bail here
		if i.DryRun {
			i.cfg.Log("WARNING: This chart or one of its subcharts contains CRDs. Rendering may fail or contain inaccuracies.")
		} else if err := i.applyCRDs(crds); err != nil {
			return nil, err
		}
	}

	if i.ClientOnly {
		// Add mock objects in here so it doesn't use Kube API server
		// NOTE(bacongobbler): used for `helm template`
		i.cfg.Capabilities = chartutil.DefaultCapabilities
		i.cfg.Capabilities.APIVersions = append(i.cfg.Capabilities.APIVersions, i.APIVersions...)
		i.cfg.KubeClient = &kubefake.PrintingKubeClient{Out: ioutil.Discard}

		mem := driver.NewMemory()
		mem.SetNamespace(i.Namespace)
		i.cfg.Releases = storage.Init(mem)
	} else if !i.ClientOnly && len(i.APIVersions) > 0 {
		i.cfg.Log("API Version list given outside of client only mode, this list will be ignored")
	}

	if err := chartutil.ProcessDependencies(chrt, vals); err != nil {
		return nil, err
	}

	// Make sure if Atomic is set, that wait is set as well. This makes it so
	// the user doesn't have to specify both
	i.Wait = i.Wait || i.Atomic

	caps, err := i.cfg.GetCapabilities()
	if err != nil {
		return nil, err
	}

	//special case for helm template --is-upgrade
	isUpgrade := i.IsUpgrade && i.DryRun
	options := chartutil.ReleaseOptions{
		Name:      i.ReleaseName,
		Namespace: i.Namespace,
		Revision:  1,
		IsInstall: !isUpgrade,
		IsUpgrade: isUpgrade,
	}
	valuesToRender, err := chartutil.ToRenderValues(chrt, vals, options, caps)
	if err != nil {
		return nil, err
	}

	rel := i.createRelease(chrt, vals)

	var manifestDoc *bytes.Buffer
	rel.Hooks, manifestDoc, rel.Info.Notes, err = i.cfg.renderResources(chrt, valuesToRender, i.ReleaseName, i.OutputDir, i.SubNotes, i.UseReleaseName, i.IncludeCRDs, i.PostRenderer, i.DryRun)
	// Even for errors, attach this if available
	if manifestDoc != nil {
		rel.Manifest = manifestDoc.String()
	}
	// Check error from render
	if err != nil {
		rel.SetStatus(release.StatusFailed, fmt.Sprintf("failed to render resource: %s", err.Error()))
		// Return a release with partial data so that the client can show debugging information.
		return rel, err
	}

	// Mark this release as in-progress
	rel.SetStatus(release.StatusPendingInstall, "Initial apply underway")

	var toBeAdopted kube.ResourceList
	resources, err := i.cfg.KubeClient.Build(bytes.NewBufferString(rel.Manifest), !i.DisableOpenAPIValidation)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build kubernetes objects from release manifest")
	}

	// It is safe to use "force" here because these are resources currently rendered by the chart.
	err = resources.Visit(setMetadataVisitor(rel.Name, rel.Namespace, true))
	if err != nil {
		return nil, err
	}

	// Apply requires an extra validation step of checking that resources
	// don't already exist before we actually create resources. If we continue
	// forward and create the release object with resources that already exist,
	// we'll end up in a state where we will delete those resources upon
	// deleting the release because the manifest will be pointing at that
	// resource
	if !i.ClientOnly && !isUpgrade && len(resources) > 0 {
		toBeAdopted, err = existingResourceConflict(resources, rel.Name, rel.Namespace)
		if err != nil {
			return nil, errors.Wrap(err, "rendered manifests contain a resource that already exists. Unable to continue with apply")
		}
	}

	// Bail out here if it is a dry run
	if i.DryRun {
		rel.Info.Description = "Dry run complete"
		return rel, nil
	}

	if i.CreateNamespace {
		ns := &v1.Namespace{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Namespace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: i.Namespace,
				Labels: map[string]string{
					"name": i.Namespace,
				},
			},
		}
		buf, err := yaml.Marshal(ns)
		if err != nil {
			return nil, err
		}
		resourceList, err := i.cfg.KubeClient.Build(bytes.NewBuffer(buf), true)
		if err != nil {
			return nil, err
		}
		if _, err := i.cfg.KubeClient.Create(resourceList); err != nil && !apierrors.IsAlreadyExists(err) {
			return nil, err
		}
	}

	// If Replace is true, we need to supercede the last release.
	if i.Replace {
		if err := i.replaceRelease(rel); err != nil {
			return nil, err
		}
	}

	// Store the release in history before continuing (new in Helm 3). We always know
	// that this is a create operation.
	if err := i.cfg.Releases.Create(rel); err != nil {
		// We could try to recover gracefully here, but since nothing has been applyed
		// yet, this is probably safer than trying to continue when we know storage is
		// not working.
		return rel, err
	}

	// pre-apply hooks
	if !i.DisableHooks {
		if err := i.cfg.execHook(rel, release.HookPreInstall, i.Timeout); err != nil {
			return i.failRelease(rel, fmt.Errorf("failed pre-apply: %s", err))
		}
	}

	// At this point, we can do the apply. Note that before we were detecting whether to
	// do an update, but it's not clear whether we WANT to do an update if the re-use is set
	// to true, since that is basically an upgrade operation.
	if len(toBeAdopted) == 0 && len(resources) > 0 {

		errs := []error{}
		for _, info := range resources {
			if err := i.ApplyOptions.ApplyOneObject(info); err != nil {
				errs = append(errs, err)
			}
		}
		// If any errors occurred during apply, then return error (or
		// aggregate of errors).
		if len(errs) == 1 {
			return i.failRelease(rel, errs[0])
		}
		if len(errs) > 1 {
			return i.failRelease(rel, utilerrors.NewAggregate(errs))
		}
		//if _, err := i.cfg.KubeClient.Create(resources); err != nil {
		//	return i.failRelease(rel, err)
		//}

	} else if len(resources) > 0 {
		if _, err := i.Update(toBeAdopted, resources, false); err != nil {
			return i.failRelease(rel, err)
		}
	}

	if i.Wait {
		if err := i.cfg.KubeClient.Wait(resources, i.Timeout); err != nil {
			return i.failRelease(rel, err)
		}

	}

	if !i.DisableHooks {
		if err := i.cfg.execHook(rel, release.HookPostInstall, i.Timeout); err != nil {
			return i.failRelease(rel, fmt.Errorf("failed post-apply: %s", err))
		}
	}

	if len(i.Description) > 0 {
		rel.SetStatus(release.StatusDeployed, i.Description)
	} else {
		rel.SetStatus(release.StatusDeployed, "Apply complete")
	}

	// This is a tricky case. The release has been created, but the result
	// cannot be recorded. The truest thing to tell the user is that the
	// release was created. However, the user will not be able to do anything
	// further with this release.
	//
	// One possible strategy would be to do a timed retry to see if we can get
	// this stored in the future.
	if err := i.recordRelease(rel); err != nil {
		i.cfg.Log("failed to record the release: %s", err)
	}

	return rel, nil
}

func (i *Apply) failRelease(rel *release.Release, err error) (*release.Release, error) {
	rel.SetStatus(release.StatusFailed, fmt.Sprintf("Release %q failed: %s", i.ReleaseName, err.Error()))
	if i.Atomic {
		i.cfg.Log("Apply failed and atomic is set, unapplying release")
		unapply := NewUninstall(i.cfg)
		unapply.DisableHooks = i.DisableHooks
		unapply.KeepHistory = false
		unapply.Timeout = i.Timeout
		if _, unapplyErr := unapply.Run(i.ReleaseName); unapplyErr != nil {
			return rel, errors.Wrapf(unapplyErr, "an error occurred while unapplying the release. original apply error: %s", err)
		}
		return rel, errors.Wrapf(err, "release %s failed, and has been unapplyed due to atomic being set", i.ReleaseName)
	}
	i.recordRelease(rel) // Ignore the error, since we have another error to deal with.
	return rel, err
}

// availableName tests whether a name is available
//
// Roughly, this will return an error if name is
//
//	- empty
//	- too long
//	- already in use, and not deleted
//	- used by a deleted release, and i.Replace is false
func (i *Apply) availableName() error {
	start := i.ReleaseName
	if start == "" {
		return errors.New("name is required")
	}

	if len(start) > releaseNameMaxLen {
		return errors.Errorf("release name %q exceeds max length of %d", start, releaseNameMaxLen)
	}

	if i.DryRun {
		return nil
	}

	h, err := i.cfg.Releases.History(start)
	if err != nil || len(h) < 1 {
		return nil
	}
	releaseutil.Reverse(h, releaseutil.SortByRevision)
	rel := h[0]

	if st := rel.Info.Status; i.Replace && (st == release.StatusUninstalled || st == release.StatusFailed) {
		return nil
	}
	return errors.New("cannot re-use a name that is still in use")
}

// createRelease creates a new release object
func (i *Apply) createRelease(chrt *chart.Chart, rawVals map[string]interface{}) *release.Release {
	ts := i.cfg.Now()
	return &release.Release{
		Name:      i.ReleaseName,
		Namespace: i.Namespace,
		Chart:     chrt,
		Config:    rawVals,
		Info: &release.Info{
			FirstDeployed: ts,
			LastDeployed:  ts,
			Status:        release.StatusUnknown,
		},
		Version: 1,
	}
}

// recordRelease with an update operation in case reuse has been set.
func (i *Apply) recordRelease(r *release.Release) error {
	// This is a legacy function which has been reduced to a oneliner. Could probably
	// refactor it out.
	return i.cfg.Releases.Update(r)
}

// replaceRelease replaces an older release with this one
//
// This allows us to re-use names by superseding an existing release with a new one
func (i *Apply) replaceRelease(rel *release.Release) error {
	hist, err := i.cfg.Releases.History(rel.Name)
	if err != nil || len(hist) == 0 {
		// No releases exist for this name, so we can return early
		return nil
	}

	releaseutil.Reverse(hist, releaseutil.SortByRevision)
	last := hist[0]

	// Update version to the next available
	rel.Version = last.Version + 1

	// Do not change the status of a failed release.
	if last.Info.Status == release.StatusFailed {
		return nil
	}

	// For any other status, mark it as superseded and store the old record
	last.SetStatus(release.StatusSuperseded, "superseded by new release")
	return i.recordRelease(last)
}

// NameAndChart returns the name and chart that should be used.
//
// This will read the flags and handle name generation if necessary.
func (i *Apply) NameAndChart(args []string) (string, string, error) {
	flagsNotSet := func() error {
		if i.GenerateName {
			return errors.New("cannot set --generate-name and also specify a name")
		}
		if i.NameTemplate != "" {
			return errors.New("cannot set --name-template and also specify a name")
		}
		return nil
	}

	if len(args) > 2 {
		return args[0], args[1], errors.Errorf("expected at most two arguments, unexpected arguments: %v", strings.Join(args[2:], ", "))
	}

	if len(args) == 2 {
		return args[0], args[1], flagsNotSet()
	}

	if i.NameTemplate != "" {
		name, err := TemplateName(i.NameTemplate)
		return name, args[0], err
	}

	if i.ReleaseName != "" {
		return i.ReleaseName, args[0], nil
	}

	if !i.GenerateName {
		return "", args[0], errors.New("must either provide a name or specify --generate-name")
	}

	base := filepath.Base(args[0])
	if base == "." || base == "" {
		base = "chart"
	}
	// if present, strip out the file extension from the name
	if idx := strings.Index(base, "."); idx != -1 {
		base = base[0:idx]
	}

	return fmt.Sprintf("%s-%d", base, time.Now().Unix()), args[0], nil
}

// ------------------------------------------------------
var metadataAccessor = meta.NewAccessor()

func (i *Apply) Update(original, target kube.ResourceList, force bool) (*kube.Result, error) {
	res := &kube.Result{}

	i.cfg.Log("checking %d resources for changes", len(target))

	errs := []error{}
	for _, info := range target {
		if err := i.ApplyOptions.ApplyOneObject(info); err != nil {
			errs = append(errs, err)
		}
	}
	// If any errors occurred during apply, then return error (or
	// aggregate of errors).
	if len(errs) == 1 {
		return nil, errs[0]
	}
	if len(errs) > 1 {
		return nil, utilerrors.NewAggregate(errs)
	}
	//if _, err := i.cfg.KubeClient.Create(resources); err != nil {
	//	return i.failRelease(rel, err)
	//}

	for _, info := range original.Difference(target) {
		i.cfg.Log("Deleting %q in %s...", info.Name, info.Namespace)

		if err := info.Get(); err != nil {
			i.cfg.Log("Unable to get obj %q, err: %s", info.Name, err)
		}
		annotations, err := metadataAccessor.Annotations(info.Object)
		if err != nil {
			i.cfg.Log("Unable to get annotations on %q, err: %s", info.Name, err)
		}
		if annotations != nil && annotations[kube.ResourcePolicyAnno] == kube.KeepPolicy {
			i.cfg.Log("Skipping delete of %q due to annotation [%s=%s]", info.Name, kube.ResourcePolicyAnno, kube.KeepPolicy)
			continue
		}

		res.Deleted = append(res.Deleted, info)
		if err := deleteResource(info); err != nil {
			if apierrors.IsNotFound(err) {
				i.cfg.Log("Attempted to delete %q, but the resource was missing", info.Name)
			} else {
				i.cfg.Log("Failed to delete %q, err: %s", info.Name, err)
				return res, errors.Wrapf(err, "Failed to delete %q", info.Name)
			}
		}
	}
	return res, nil
}

func deleteResource(info *resource.Info) error {
	policy := metav1.DeletePropagationBackground
	opts := &metav1.DeleteOptions{PropagationPolicy: &policy}
	_, err := resource.NewHelper(info.Client, info.Mapping).DeleteWithOptions(info.Namespace, info.Name, opts)
	return err
}
