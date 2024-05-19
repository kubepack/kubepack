package repo

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	fluxsrc "github.com/fluxcd/source-controller/api/v1"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmgetter "helm.sh/helm/v3/pkg/getter"
	helmreg "helm.sh/helm/v3/pkg/registry"
	"k8s.io/klog/v2"
	"kubepack.dev/lib-helm/pkg/internal/helm/getter"
	"kubepack.dev/lib-helm/pkg/internal/helm/registry"
	"kubepack.dev/lib-helm/pkg/internal/helm/repository"
	"sigs.k8s.io/controller-runtime/pkg/client"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

var getters = helmgetter.Providers{
	helmgetter.Provider{
		Schemes: []string{"http", "https"},
		New:     helmgetter.NewHTTPGetter,
	},
	helmgetter.Provider{
		Schemes: []string{"oci"},
		New:     helmgetter.NewOCIGetter,
	},
}

func (r *Registry) getFluxChart(obj releasesapi.ChartSourceRef) (*ChartExtended, error) {
	ctx := context.TODO()

	var repo fluxsrc.HelmRepository
	err := r.kc.Get(ctx, client.ObjectKey{Namespace: obj.SourceRef.Namespace, Name: obj.SourceRef.Name}, &repo)
	if err != nil {
		return nil, err
	}

	switch repo.Spec.Type {
	case fluxsrc.HelmRepositoryTypeOCI:
		// Used to login with the repository declared provider
		ctxTimeout, cancel := context.WithTimeout(ctx, repo.GetTimeout())
		defer cancel()

		normalizedURL, err := repository.NormalizeURL(repo.Spec.URL)
		if err != nil {
			return nil, chartRepoConfigErrorReturn(err)
		}

		clientOpts, certsTmpDir, err := getter.GetClientOpts(ctxTimeout, r.kc, &repo, normalizedURL)
		if err != nil && !errors.Is(err, getter.ErrDeprecatedTLSConfig) {
			return nil, err
		}
		if certsTmpDir != "" {
			defer func() {
				if err := os.RemoveAll(certsTmpDir); err != nil {
					klog.Warningf("failed to delete temporary certificates directory: %s", err)
				}
			}()
		}

		getterOpts := clientOpts.GetterOpts

		// Initialize the chart repository
		var chartRepo repository.Downloader
		if !helmreg.IsOCI(normalizedURL) {
			return nil, fmt.Errorf("invalid OCI registry URL: %s", normalizedURL)
		}

		// with this function call, we create a temporary file to store the credentials if needed.
		// this is needed because otherwise the credentials are stored in ~/.docker/config.json.
		// TODO@souleb: remove this once the registry move to Oras v2
		// or rework to enable reusing credentials to avoid the unneccessary handshake operations
		registryClient, credentialsFile, err := registry.ClientGenerator(clientOpts.TlsConfig, clientOpts.MustLoginToRegistry())
		if err != nil {
			return nil, fmt.Errorf("failed to construct Helm client: %w", err)
		}

		if credentialsFile != "" {
			defer func() {
				if err := os.Remove(credentialsFile); err != nil {
					klog.Warningf("failed to delete temporary credentials file: %s", err)
				}
			}()
		}

		/*
			// TODO(tamal): SKIP verifier
			var verifiers []soci.Verifier
			if obj.Spec.Verify != nil {
				provider := obj.Spec.Verify.Provider
				verifiers, err = r.makeVerifiers(ctx, obj, *clientOpts)
				if err != nil {
					if obj.Spec.Verify.SecretRef == nil {
						provider = fmt.Sprintf("%s keyless", provider)
					}
					e := &serror.Event{
						Err:    fmt.Errorf("failed to verify the signature using provider '%s': %w", provider, err),
						Reason: sourcev1.VerificationError,
					}
					conditions.MarkFalse(obj, sourcev1.SourceVerifiedCondition, e.Reason, e.Err.Error())
					return sreconcile.ResultEmpty, e
				}
			}
		*/

		// Tell the chart repository to use the OCI client with the configured getter
		getterOpts = append(getterOpts, helmgetter.WithRegistryClient(registryClient))
		ociChartRepo, err := repository.NewOCIChartRepository(normalizedURL,
			repository.WithOCIGetter(getters),
			repository.WithOCIGetterOptions(getterOpts),
			repository.WithOCIRegistryClient(registryClient),
			// repository.WithVerifiers(verifiers),
		)
		if err != nil {
			return nil, chartRepoConfigErrorReturn(err)
		}

		// If login options are configured, use them to login to the registry
		// The OCIGetter will later retrieve the stored credentials to pull the chart
		if clientOpts.MustLoginToRegistry() {
			err = ociChartRepo.Login(clientOpts.RegLoginOpts...)
			if err != nil {
				return nil, fmt.Errorf("failed to login to OCI registry: %w", err)
			}
			defer ociChartRepo.Logout()
		}
		chartRepo = ociChartRepo

		// https://github.com/fluxcd/source-controller/blob/04d87b61ca76e8081869cf3f9937bc178195f876/controllers/helmchart_controller.go#L467
		remote := chartRepo
		// Get the current version for the RemoteReference
		cv, err := remote.GetChartVersion(obj.Name, obj.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to get chart version for remote reference: %w", err)
		}
		// Download the package for the resolved version
		res, err := remote.DownloadChart(cv)
		if err != nil {
			return nil, fmt.Errorf("failed to download chart for remote reference: %w", err)
		}
		chrt, err := loader.LoadArchive(res)
		if err != nil {
			return nil, err
		}
		return &ChartExtended{Chart: chrt}, nil
	default:
		repositoryUrl, err := r.registerHelmRepository(repo)
		if err != nil {
			return nil, err
		}
		return r.getLegacyChart(repositoryUrl, obj.Name, obj.Version)
	}
}

func chartRepoConfigErrorReturn(err error) error {
	switch err.(type) {
	case *url.Error:
		return fmt.Errorf("invalid Helm repository URL: %w", err)
	default:
		return fmt.Errorf("failed to construct Helm client: %w", err)
	}
}
