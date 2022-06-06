module gitlab.com/project-emco/core/emco-base/src/workflowmgr

require (
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/klauspost/compress v1.15.0 // indirect
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.18.1
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	gitlab.com/project-emco/core/emco-base/src/orchestrator v0.0.0-20220308165843-684035a6c4f0
	go.mongodb.org/mongo-driver v1.8.4 // indirect
	go.temporal.io/api v1.7.0
	go.temporal.io/sdk v1.13.1
	golang.org/x/crypto v0.0.0-20220307211146-efcb8507fb70 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220307203707-22a9840ba4d7 // indirect
	golang.org/x/time v0.0.0-20220224211638-0e9765cccd65 // indirect
	google.golang.org/genproto v0.0.0-20220308174144-ae0e22291548 // indirect
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.7.0
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	gitlab.com/project-emco/core/emco-base/src/clm => ../clm
	gitlab.com/project-emco/core/emco-base/src/monitor => ../monitor
	gitlab.com/project-emco/core/emco-base/src/orchestrator => ../orchestrator
	gitlab.com/project-emco/core/emco-base/src/rsync => ../rsync
	gitlab.com/project-emco/core/emco-base/src/workflowmgr => ../workflowmgr
	gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/emcotemporalapi => ../workflowmgr/pkg/emcotemporalapi
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200819165624-17cef6e3e9d5 // 17cef6e3e9d5 is the SHA for git tag v3.4.12
	helm.sh/helm/v3 => helm.sh/helm/v3 v3.5.3
	k8s.io/api => k8s.io/api v0.23.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.23.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.23.3
	k8s.io/apiserver => k8s.io/apiserver v0.23.3
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.23.3
	k8s.io/client-go => k8s.io/client-go v0.23.3
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.23.3
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.23.3
	k8s.io/code-generator => k8s.io/code-generator v0.23.3
	k8s.io/component-base => k8s.io/component-base v0.23.3
	k8s.io/cri-api => k8s.io/cri-api v0.23.3
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.23.3
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.23.3
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.23.3
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.23.3
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.23.3
	k8s.io/kubectl => k8s.io/kubectl v0.23.3
	k8s.io/kubelet => k8s.io/kubelet v0.23.3
	k8s.io/kubernetes => github.com/kubernetes/kubernetes v1.19.4
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.23.3
	k8s.io/metrics => k8s.io/metrics v0.23.3
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.23.3
)

go 1.16
