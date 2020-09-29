module kubepack.dev/kubepack

go 1.12

require (
	github.com/Azure/go-autorest/autorest v0.10.2 // indirect
	github.com/appscode/go v0.0.0-20200928211031-cc0c23082d91
	github.com/appscode/static-assets v0.4.1
	github.com/coreos/prometheus-operator v0.39.0
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/fatih/color v1.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.1.1
	github.com/go-macaron/binding v1.1.0
	github.com/go-openapi/spec v0.19.8
	github.com/go-openapi/swag v0.19.9 // indirect
	github.com/gobuffalo/flect v0.2.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/gofuzz v1.1.0
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/gophercloud/gophercloud v0.11.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7
	github.com/grpc-ecosystem/grpc-gateway v1.14.6 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mitchellh/mapstructure v1.2.2 // indirect
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.6.0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stripe/stripe-go v70.15.0+incompatible
	go.uber.org/atomic v1.6.0 // indirect
	go.uber.org/zap v1.13.0 // indirect
	gocloud.dev v0.20.0
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	gomodules.xyz/blobfs v0.1.2
	gomodules.xyz/jsonpatch/v2 v2.1.0
	gomodules.xyz/version v0.1.0
	gopkg.in/macaron.v1 v1.3.8
	gopkg.in/yaml.v2 v2.3.0 // indirect
	helm.sh/helm/v3 v3.2.1
	k8s.io/api v0.18.9
	k8s.io/apiextensions-apiserver v0.18.9
	k8s.io/apimachinery v0.18.9
	k8s.io/apiserver v0.18.9
	k8s.io/cli-runtime v0.18.9
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	k8s.io/kubernetes v1.18.9
	k8s.io/utils v0.0.0-20200414100711-2df71ebbae66 // indirect
	kmodules.xyz/client-go v0.0.0-20200929030759-cce6a3c623c1
	kmodules.xyz/crd-schema-fuzz v0.0.0-20200922204806-c1426cd7fcf4
	kmodules.xyz/custom-resources v0.0.0-20200922210108-70f2815a43bb
	kmodules.xyz/resource-metadata v0.4.2
	kmodules.xyz/webhook-runtime v0.0.0-20200922211931-8337935590de
	kubepack.dev/lib-helm v0.2.0
	sigs.k8s.io/application v0.8.2-0.20200306235134-f10d9ca8abd4
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/application => github.com/kubepack/application v0.8.4-0.20200705202912-9d241d6484e3

replace helm.sh/helm/v3 => github.com/kubepack/helm/v3 v3.2.2-0.20200523120511-a86fc03a6a93

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

replace k8s.io/api => github.com/kmodules/api v0.18.10-0.20200922195318-d60fe725dea0

replace k8s.io/apimachinery => github.com/kmodules/apimachinery v0.19.0-alpha.0.0.20200922195535-0c9a1b86beec

replace k8s.io/apiserver => github.com/kmodules/apiserver v0.18.10-0.20200922195747-1bd1cc8f00d1

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.9

replace k8s.io/client-go => github.com/kmodules/k8s-client-go v0.18.10-0.20200922201634-73fedf3d677e

replace k8s.io/component-base => k8s.io/component-base v0.18.9

replace k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6

replace k8s.io/kubernetes => github.com/kmodules/kubernetes v1.19.0-alpha.0.0.20200922200158-8b13196d8dc4

replace k8s.io/utils => k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89
