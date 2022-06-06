module gitlab.com/project-emco/core/emco-base/examples/sample-controller

require (
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.18.1
	github.com/pkg/errors v0.9.1
	gitlab.com/project-emco/core/emco-base/src/orchestrator v0.0.0-00010101000000-000000000000
	gitlab.com/project-emco/core/emco-base/src/rsync v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.43.0
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	gitlab.com/project-emco/core/emco-base/src/clm => ../../../../src/clm //Please modify accordingly.
	gitlab.com/project-emco/core/emco-base/src/monitor => ../../../../src/monitor //Please modify accordingly.
	gitlab.com/project-emco/core/emco-base/src/orchestrator => ../../../../src/orchestrator //Please modify accordingly.
	gitlab.com/project-emco/core/emco-base/src/rsync => ../../../../src/rsync //Please modify accordingly.
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200819165624-17cef6e3e9d5 // 17cef6e3e9d5 is the SHA for git tag v3.4.12
	google.golang.org/grpc => google.golang.org/grpc v1.28.0
	helm.sh/helm/v3 => helm.sh/helm/v3 v3.5.3
)

go 1.16
