module kubepack.dev/kubepack

go 1.12

require (
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.2 // indirect
	github.com/Unknwon/com v0.0.0-20190321035513-0fed4efef755 // indirect
	github.com/appscode/go v0.0.0-20200225060711-86360b91102a
	github.com/appscode/static-assets v0.4.1
	github.com/aws/aws-sdk-go v1.20.20 // indirect
	github.com/coreos/prometheus-operator v0.34.0
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/gabriel-vasile/mimetype v1.0.2
	github.com/go-macaron/binding v0.0.0-00010101000000-000000000000
	github.com/go-macaron/inject v0.0.0-20160627170012-d8a0b8677191 // indirect
	github.com/go-openapi/spec v0.19.4
	github.com/gobuffalo/flect v0.1.7
	github.com/gogo/protobuf v1.3.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/gofuzz v1.0.0
	github.com/google/uuid v1.1.1
	github.com/gorilla/schema v1.1.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stripe/stripe-go v68.20.0+incompatible
	gocloud.dev v0.18.0
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	gomodules.xyz/jsonpatch/v2 v2.1.0
	gomodules.xyz/version v0.1.0
	gopkg.in/ini.v1 v1.52.0 // indirect
	gopkg.in/macaron.v1 v1.3.2
	helm.sh/helm/v3 v3.0.3
	k8s.io/api v0.0.0-20191114100352-16d7abae0d2a
	k8s.io/apiextensions-apiserver v0.0.0-20191114105449-027877536833
	k8s.io/apimachinery v0.16.5-beta.1
	k8s.io/apiserver v0.0.0-20191114103151-9ca1dc586682
	k8s.io/cli-runtime v0.0.0-20191114110141-0a35778df828
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20190918143330-0270cf2f1c1d
	k8s.io/kubernetes v1.16.3
	kmodules.xyz/client-go v0.0.0-20200216080917-08714f78f885
	kmodules.xyz/crd-schema-fuzz v0.0.0-20191129174258-81f984340891
	kmodules.xyz/custom-resources v0.0.0-20191130062942-f41b54f62419
	kmodules.xyz/resource-metadata v0.3.10
	kmodules.xyz/webhook-runtime v0.0.0-20191127075323-d4bfdee6974d
	kubepack.dev/lib-helm v0.0.0-20200430114938-bb329c6b34fe
	sigs.k8s.io/yaml v1.1.0
)

replace (
	cloud.google.com/go => cloud.google.com/go v0.38.0
	git.apache.org/thrift.git => github.com/apache/thrift v0.13.0
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v32.5.0+incompatible
	github.com/Azure/go-ansiterm => github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.9.0
	github.com/Azure/go-autorest/autorest/adal => github.com/Azure/go-autorest/autorest/adal v0.5.0
	github.com/Azure/go-autorest/autorest/date => github.com/Azure/go-autorest/autorest/date v0.1.0
	github.com/Azure/go-autorest/autorest/mocks => github.com/Azure/go-autorest/autorest/mocks v0.2.0
	github.com/Azure/go-autorest/autorest/to => github.com/Azure/go-autorest/autorest/to v0.2.0
	github.com/Azure/go-autorest/autorest/validation => github.com/Azure/go-autorest/autorest/validation v0.1.0
	github.com/Azure/go-autorest/logger => github.com/Azure/go-autorest/logger v0.1.0
	github.com/Azure/go-autorest/tracing => github.com/Azure/go-autorest/tracing v0.5.0
	github.com/docker/docker => github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
	github.com/go-macaron/binding => github.com/gomodules/binding v0.0.0-20200226114658-71565367f820
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.1
	github.com/kubernetes-csi/external-snapshotter => github.com/kmodules/external-snapshotter v1.2.1-0.20191128100451-0265c5fa679a
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	google.golang.org/api => google.golang.org/api v0.6.1-0.20190607001116-5213b8090861
	google.golang.org/grpc => github.com/gomodules/grpc-go v1.23.2-0.20191111130652-202dbf267fb7
	helm.sh/helm/v3 => github.com/kubepack/helm/v3 v3.0.3-0.20200119202455-afb1ef54d569
	k8s.io/api => k8s.io/api v0.0.0-20191114100352-16d7abae0d2a
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191114105449-027877536833
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.0.0-20191119091232-0553326db082
	k8s.io/apiserver => github.com/kmodules/apiserver v0.0.0-20191119111000-36ac3646ae82
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191114110141-0a35778df828
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191114101535-6c5935290e33
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191114112024-4bbba8331835
	k8s.io/component-base => k8s.io/component-base v0.0.0-20191114102325-35a9586014f7
	k8s.io/klog => k8s.io/klog v0.4.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191114103820-f023614fb9ea
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191114113550-6123e1c827f7
	k8s.io/kubernetes => github.com/kmodules/kubernetes v1.17.0-alpha.0.0.20191127022853-9d027e3886fd
	k8s.io/metrics => k8s.io/metrics v0.0.0-20191114105837-a4a2842dc51b
	k8s.io/repo-infra => k8s.io/repo-infra v0.0.0-20181204233714-00fe14e3d1a3
	k8s.io/utils => k8s.io/utils v0.0.0-20190801114015-581e00157fb1
	sigs.k8s.io/kustomize => sigs.k8s.io/kustomize v2.0.3+incompatible
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v0.0.0-20190817042607-6149e4549fca
	sigs.k8s.io/yaml => sigs.k8s.io/yaml v1.1.0
)
