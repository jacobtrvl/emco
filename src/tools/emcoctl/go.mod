module tools/emcoctl

go 1.16

require (
	github.com/go-resty/resty/v2 v2.7.0
	github.com/google/uuid v1.3.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.3
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.3.0
	github.com/spf13/viper v1.10.1
	gitlab.com/project-emco/core/emco-base/src/orchestrator v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.43.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v3 v3.0.1
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	gitlab.com/project-emco/core/emco-base/src/clm => ../../clm
	gitlab.com/project-emco/core/emco-base/src/monitor => ../../monitor
	gitlab.com/project-emco/core/emco-base/src/orchestrator => ../../orchestrator
	gitlab.com/project-emco/core/emco-base/src/rsync => ../../rsync
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200819165624-17cef6e3e9d5 // 17cef6e3e9d5 is the SHA for git tag v3.4.12
	google.golang.org/grpc => google.golang.org/grpc v1.29.0
)
