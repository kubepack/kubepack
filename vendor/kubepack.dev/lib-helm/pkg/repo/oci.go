package repo

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fluxcd/pkg/oci"
	"github.com/fluxcd/pkg/oci/auth/login"
	fluxsrc "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmgetter "helm.sh/helm/v3/pkg/getter"
	helmreg "helm.sh/helm/v3/pkg/registry"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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
		var (
			// tlsConfig     *tls.Config
			authenticator authn.Authenticator
			keychain      authn.Keychain
		)
		// Used to login with the repository declared provider
		ctxTimeout, cancel := context.WithTimeout(ctx, repo.Spec.Timeout.Duration)
		defer cancel()

		normalizedURL := repository.NormalizeURL(repo.Spec.URL)
		err = repository.ValidateDepURL(normalizedURL)
		if err != nil {
			return nil, err
		}
		// Construct the Getter options from the HelmRepository data
		clientOpts := []helmgetter.Option{
			helmgetter.WithURL(normalizedURL),
			helmgetter.WithTimeout(repo.Spec.Timeout.Duration),
			helmgetter.WithPassCredentialsAll(repo.Spec.PassCredentials),
		}

		if secret, err := getHelmRepositorySecret(ctx, r.kc, &repo); secret != nil || err != nil {
			if err != nil {
				return nil, fmt.Errorf("failed to get secret '%s': %w", repo.Spec.SecretRef.Name, err)
			}

			// Build client options from secret
			opts, _, err := clientOptionsFromSecret(secret, normalizedURL)
			if err != nil {
				return nil, err
			}
			clientOpts = append(clientOpts, opts...)
			// TODO(tamal): TLS not used
			// tlsConfig = tls

			// Build registryClient options from secret
			keychain, err = registry.LoginOptionFromSecret(normalizedURL, *secret)
			if err != nil {
				return nil, fmt.Errorf("failed to configure Helm client with secret data: %w", err)
			}
		} else if repo.Spec.Provider != fluxsrc.GenericOCIProvider && repo.Spec.Type == fluxsrc.HelmRepositoryTypeOCI {
			auth, authErr := oidcAuth(ctxTimeout, repo.Spec.URL, repo.Spec.Provider)
			if authErr != nil && !errors.Is(authErr, oci.ErrUnconfiguredProvider) {
				return nil, fmt.Errorf("failed to get credential from %s: %w", repo.Spec.Provider, authErr)
			}
			if auth != nil {
				authenticator = auth
			}
		}

		loginOpt, err := makeLoginOption(authenticator, keychain, normalizedURL)
		if err != nil {
			return nil, err
		}

		// Initialize the chart repository
		var chartRepo repository.Downloader

		if !helmreg.IsOCI(normalizedURL) {
			return nil, fmt.Errorf("invalid OCI registry URL: %s", normalizedURL)
		}

		// with this function call, we create a temporary file to store the credentials if needed.
		// this is needed because otherwise the credentials are stored in ~/.docker/config.json.
		// TODO@souleb: remove this once the registry move to Oras v2
		// or rework to enable reusing credentials to avoid the unneccessary handshake operations
		registryClient, credentialsFile, err := registry.ClientGenerator(loginOpt != nil)
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
				verifiers, err = r.makeVerifiers(ctx, obj, authenticator, keychain)
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
		clientOpts = append(clientOpts, helmgetter.WithRegistryClient(registryClient))
		ociChartRepo, err := repository.NewOCIChartRepository(normalizedURL,
			repository.WithOCIGetter(getters),
			repository.WithOCIGetterOptions(clientOpts),
			repository.WithOCIRegistryClient(registryClient),
			// repository.WithVerifiers(verifiers),
		)
		if err != nil {
			return nil, err
		}
		chartRepo = ociChartRepo

		// If login options are configured, use them to login to the registry
		// The OCIGetter will later retrieve the stored credentials to pull the chart
		if loginOpt != nil {
			err = ociChartRepo.Login(loginOpt)
			if err != nil {
				return nil, fmt.Errorf("failed to login to OCI registry: %w", err)
			}
			defer ociChartRepo.Logout()
		}

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
		_, err := r.registerHelmRepository(repo)
		if err != nil {
			return nil, err
		}
		return r.getLegacyChart(obj.SourceRef.Name, obj.Name, obj.Version)
	}
}

func getHelmRepositorySecret(ctx context.Context, client client.Reader, repository *fluxsrc.HelmRepository) (*core.Secret, error) {
	if repository.Spec.SecretRef == nil {
		return nil, nil
	}
	key := types.NamespacedName{
		Namespace: repository.GetNamespace(),
		Name:      repository.Spec.SecretRef.Name,
	}
	var secret core.Secret
	err := client.Get(ctx, key, &secret)
	if err != nil {
		return nil, err
	}
	return &secret, nil
}

func clientOptionsFromSecret(secret *core.Secret, normalizedURL string) ([]helmgetter.Option, *tls.Config, error) {
	opts, err := getter.ClientOptionsFromSecret(*secret)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to configure Helm client with secret data: %w", err)
	}

	tlsConfig, err := getter.TLSClientConfigFromSecret(*secret, normalizedURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TLS client config with secret data: %w", err)
	}

	return opts, tlsConfig, nil
}

// makeLoginOption returns a registry login option for the given HelmRepository.
// If the HelmRepository does not specify a secretRef, a nil login option is returned.
func makeLoginOption(auth authn.Authenticator, keychain authn.Keychain, registryURL string) (helmreg.LoginOption, error) {
	if auth != nil {
		return registry.AuthAdaptHelper(auth)
	}

	if keychain != nil {
		return registry.KeychainAdaptHelper(keychain)(registryURL)
	}

	return nil, nil
}

// oidcAuth generates the OIDC credential authenticator based on the specified cloud provider.
func oidcAuth(ctx context.Context, url, provider string) (authn.Authenticator, error) {
	u := strings.TrimPrefix(url, fluxsrc.OCIRepositoryPrefix)
	ref, err := name.ParseReference(u)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL '%s': %w", u, err)
	}

	opts := login.ProviderOptions{}
	switch provider {
	case fluxsrc.AmazonOCIProvider:
		opts.AwsAutoLogin = true
	case fluxsrc.AzureOCIProvider:
		opts.AzureAutoLogin = true
	case fluxsrc.GoogleOCIProvider:
		opts.GcpAutoLogin = true
	}

	return login.NewManager().Login(ctx, u, ref, opts)
}
