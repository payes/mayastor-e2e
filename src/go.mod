module mayastor-e2e

go 1.15

require (
	github.com/container-storage-interface/spec v1.2.0 // indirect
	github.com/go-openapi/errors v0.19.9
	github.com/go-openapi/loads v0.20.2
	github.com/go-openapi/runtime v0.19.24
	github.com/go-openapi/spec v0.20.3
	github.com/go-openapi/strfmt v0.20.0
	github.com/go-openapi/swag v0.19.14
	github.com/go-openapi/validate v0.20.2
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.7.3
	github.com/ilyakaznacheev/cleanenv v1.2.5
	github.com/jessevdk/go-flags v1.5.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/pkg/errors v0.9.1
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6
	google.golang.org/grpc v1.37.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/klog/v2 v2.4.0
	k8s.io/kubernetes v1.19.0
	sigs.k8s.io/controller-runtime v0.7.0
)

replace k8s.io/api => k8s.io/api v0.19.0

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.0

replace k8s.io/apimachinery => k8s.io/apimachinery v0.19.0

replace k8s.io/apiserver => k8s.io/apiserver v0.19.0

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.0

replace k8s.io/client-go => k8s.io/client-go v0.19.0

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.19.0

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.19.0

replace k8s.io/code-generator => k8s.io/code-generator v0.19.0

replace k8s.io/component-base => k8s.io/component-base v0.19.0

replace k8s.io/cri-api => k8s.io/cri-api v0.19.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.19.0

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.19.0

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.19.0

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.19.0

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.19.0

replace k8s.io/kubectl => k8s.io/kubectl v0.19.0

replace k8s.io/kubelet => k8s.io/kubelet v0.19.0

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.19.0

replace k8s.io/metrics => k8s.io/metrics v0.19.0

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.19.0

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.19.0

replace k8s.io/sample-controller => k8s.io/sample-controller v0.19.0
