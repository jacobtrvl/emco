module gitlab.com/project-emco/core/emco-base/src/rsync

go 1.16

require (
	github.com/fluxcd/go-git-providers v0.5.4
	github.com/fluxcd/kustomize-controller/api v0.21.1
	github.com/fluxcd/source-controller/api v0.21.2
	github.com/ghodss/yaml v1.0.0
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/go-github/v41 v41.0.0
	github.com/jonboulle/clockwork v0.2.2
	github.com/pkg/errors v0.9.1
	gitlab.com/project-emco/core/emco-base/src/monitor v0.0.0-00010101000000-000000000000
	gitlab.com/project-emco/core/emco-base/src/orchestrator v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/grpc v1.43.0
	google.golang.org/protobuf v1.27.1
	k8s.io/api v0.23.3
	k8s.io/apiextensions-apiserver v0.23.3
	k8s.io/apimachinery v0.23.3
	k8s.io/cli-runtime v0.23.3
	k8s.io/client-go v0.23.3
	k8s.io/kube-openapi v0.0.0-20220124234850-424119656bbf
	k8s.io/kubectl v0.23.3
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	gitlab.com/project-emco/core/emco-base/src/clm => ../clm
	gitlab.com/project-emco/core/emco-base/src/monitor => ../monitor
	gitlab.com/project-emco/core/emco-base/src/orchestrator => ../orchestrator
	gitlab.com/project-emco/core/emco-base/src/rsync => ../rsync
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200819165624-17cef6e3e9d5 // 17cef6e3e9d5 is the SHA for git tag v3.4.12
	google.golang.org/grpc => google.golang.org/grpc v1.29.0
)
