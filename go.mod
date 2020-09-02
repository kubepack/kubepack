module kubepack.dev/kubepack

go 1.12

require (
	github.com/appscode/go v0.0.0-20200323182826-54e98e09185a
	github.com/appscode/static-assets v0.4.1
	github.com/coreos/prometheus-operator v0.39.0
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/gabriel-vasile/mimetype v1.1.0
	github.com/go-macaron/binding v1.1.0
	github.com/go-openapi/spec v0.19.3
	github.com/gobuffalo/flect v0.2.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/gofuzz v1.1.0
	github.com/google/uuid v1.1.1
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stripe/stripe-go v70.15.0+incompatible
	gocloud.dev v0.20.0
	gomodules.xyz/blobfs v0.1.2
	gomodules.xyz/jsonpatch/v2 v2.1.0
	gomodules.xyz/version v0.1.0
	gopkg.in/macaron.v1 v1.3.8
	helm.sh/helm/v3 v3.2.1
	k8s.io/api v0.18.3
	k8s.io/apiextensions-apiserver v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/apiserver v0.18.3
	k8s.io/cli-runtime v0.18.3
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	k8s.io/kubernetes v1.18.3
	kmodules.xyz/client-go v0.0.0-20200901064306-0f1faee534af
	kmodules.xyz/crd-schema-fuzz v0.0.0-20200521005638-2433a187de95
	kmodules.xyz/custom-resources v0.0.0-20200604135349-9e9f5c4fdba9
	kmodules.xyz/resource-metadata v0.4.2
	kmodules.xyz/webhook-runtime v0.0.0-20200522123600-ca70a7e28ed0
	kubepack.dev/lib-helm v0.2.0
	sigs.k8s.io/application v0.8.2-0.20200306235134-f10d9ca8abd4
	sigs.k8s.io/yaml v1.2.0
)

replace google.golang.org/api => google.golang.org/api v0.14.0

replace google.golang.org/genproto => google.golang.org/genproto v0.0.0-20191115194625-c23dd37a84c9

replace cloud.google.com/go => cloud.google.com/go v0.49.0

replace sigs.k8s.io/application => github.com/kubepack/application v0.8.4-0.20200705202912-9d241d6484e3

replace helm.sh/helm/v3 => github.com/kubepack/helm/v3 v3.2.2-0.20200523120511-a86fc03a6a93

replace bitbucket.org/ww/goautoneg => gomodules.xyz/goautoneg v0.0.0-20120707110453-a547fc61f48d

replace git.apache.org/thrift.git => github.com/apache/thrift v0.13.0

replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v35.0.0+incompatible

replace github.com/Azure/go-ansiterm => github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible

replace github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.9.0

replace github.com/Azure/go-autorest/autorest/adal => github.com/Azure/go-autorest/autorest/adal v0.5.0

replace github.com/Azure/go-autorest/autorest/azure/auth => github.com/Azure/go-autorest/autorest/azure/auth v0.2.0

replace github.com/Azure/go-autorest/autorest/date => github.com/Azure/go-autorest/autorest/date v0.1.0

replace github.com/Azure/go-autorest/autorest/mocks => github.com/Azure/go-autorest/autorest/mocks v0.2.0

replace github.com/Azure/go-autorest/autorest/to => github.com/Azure/go-autorest/autorest/to v0.2.0

replace github.com/Azure/go-autorest/autorest/validation => github.com/Azure/go-autorest/autorest/validation v0.1.0

replace github.com/Azure/go-autorest/logger => github.com/Azure/go-autorest/logger v0.1.0

replace github.com/Azure/go-autorest/tracing => github.com/Azure/go-autorest/tracing v0.5.0

replace github.com/imdario/mergo => github.com/imdario/mergo v0.3.5

replace github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.0.0

replace go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

replace k8s.io/api => github.com/kmodules/api v0.18.4-0.20200524125823-c8bc107809b9

replace k8s.io/apimachinery => github.com/kmodules/apimachinery v0.19.0-alpha.0.0.20200520235721-10b58e57a423

replace k8s.io/apiserver => github.com/kmodules/apiserver v0.18.4-0.20200521000930-14c5f6df9625

replace k8s.io/client-go => k8s.io/client-go v0.18.3

replace k8s.io/kubernetes => github.com/kmodules/kubernetes v1.19.0-alpha.0.0.20200521033432-49d3646051ad
