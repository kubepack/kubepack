module kubepack.dev/kubepack

go 1.12

require (
	github.com/appscode/static-assets v0.4.1
	github.com/coreos/prometheus-operator v0.39.0
	github.com/cubewise-code/go-mime v0.0.0-20200519001935-8c5762b177d8
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/gabriel-vasile/mimetype v1.1.1
	github.com/go-macaron/binding v1.1.0
	github.com/go-openapi/spec v0.19.8
	github.com/gobuffalo/flect v0.2.2
	github.com/gogo/protobuf v1.3.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/gofuzz v1.1.0
	github.com/google/uuid v1.1.2
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stripe/stripe-go v70.15.0+incompatible
	gocloud.dev v0.20.0
	gomodules.xyz/blobfs v0.1.2
	gomodules.xyz/jsonpatch/v2 v2.1.0
	gomodules.xyz/jsonpatch/v3 v3.0.1
	gomodules.xyz/runtime v0.0.0-20201104200926-d838b09dda8b
	gomodules.xyz/version v0.1.0
	gomodules.xyz/x v0.0.0-20201105065653-91c568df6331
	gopkg.in/macaron.v1 v1.3.8
	helm.sh/helm/v3 v3.4.1
	k8s.io/api v0.18.9
	k8s.io/apiextensions-apiserver v0.18.9
	k8s.io/apimachinery v0.18.9
	k8s.io/apiserver v0.18.9
	k8s.io/cli-runtime v0.18.9
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/component-base v0.18.9
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	k8s.io/kubectl v0.18.9
	k8s.io/kubernetes v1.18.9
	kmodules.xyz/client-go v0.0.0-20201230092550-8ca15cfcbefa
	kmodules.xyz/crd-schema-fuzz v0.0.0-20200922204806-c1426cd7fcf4
	kmodules.xyz/custom-resources v0.0.0-20201124062543-bd8d35c21b0c
	kmodules.xyz/resource-metadata v0.4.8-0.20210109211859-ee04a80b905a
	kmodules.xyz/webhook-runtime v0.0.0-20201105073856-2dc7382b88c6
	kubepack.dev/cli v0.0.0-20210112035115-09fc3b76e2ea
	kubepack.dev/lib-helm v0.2.1
	sigs.k8s.io/application v0.8.2-0.20200306235134-f10d9ca8abd4
	sigs.k8s.io/yaml v1.2.0
)

replace bitbucket.org/ww/goautoneg => gomodules.xyz/goautoneg v0.0.0-20120707110453-a547fc61f48d

replace cloud.google.com/go => cloud.google.com/go v0.49.0

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

replace github.com/go-macaron/binding => github.com/gomodules/binding v0.0.0-20200811095614-c752727d2156

replace github.com/go-openapi/analysis => github.com/go-openapi/analysis v0.19.5

replace github.com/go-openapi/errors => github.com/go-openapi/errors v0.19.2

replace github.com/go-openapi/jsonpointer => github.com/go-openapi/jsonpointer v0.19.3

replace github.com/go-openapi/jsonreference => github.com/go-openapi/jsonreference v0.19.3

replace github.com/go-openapi/loads => github.com/go-openapi/loads v0.19.4

replace github.com/go-openapi/runtime => github.com/go-openapi/runtime v0.19.4

replace github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.3

replace github.com/go-openapi/strfmt => github.com/go-openapi/strfmt v0.19.3

replace github.com/go-openapi/swag => github.com/go-openapi/swag v0.19.5

replace github.com/go-openapi/validate => github.com/go-openapi/validate v0.19.5

replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.1

replace github.com/golang/protobuf => github.com/golang/protobuf v1.3.2

replace github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1

replace github.com/imdario/mergo => github.com/imdario/mergo v0.3.5

replace github.com/prometheus-operator/prometheus-operator => github.com/prometheus-operator/prometheus-operator v0.42.0

replace github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring => github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.42.0

replace github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.7.1

replace go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738

replace google.golang.org/api => google.golang.org/api v0.14.0

replace google.golang.org/genproto => google.golang.org/genproto v0.0.0-20191115194625-c23dd37a84c9

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

replace helm.sh/helm/v3 => github.com/kubepack/helm/v3 v3.2.2-0.20200523120511-a86fc03a6a93

replace k8s.io/api => github.com/kmodules/api v0.18.10-0.20200922195318-d60fe725dea0

replace k8s.io/apimachinery => github.com/kmodules/apimachinery v0.19.0-alpha.0.0.20200922195535-0c9a1b86beec

replace k8s.io/apiserver => github.com/kmodules/apiserver v0.18.10-0.20200922195747-1bd1cc8f00d1

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.9

replace k8s.io/client-go => github.com/kmodules/k8s-client-go v0.18.10-0.20200922201634-73fedf3d677e

replace k8s.io/component-base => k8s.io/component-base v0.18.9

replace k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6

replace k8s.io/kubernetes => github.com/kmodules/kubernetes v1.19.0-alpha.0.0.20200922200158-8b13196d8dc4

replace k8s.io/utils => k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89

replace sigs.k8s.io/application => github.com/kubepack/application v0.8.4-0.20201117013009-57cb1e10e2ed
