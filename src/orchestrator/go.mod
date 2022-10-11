module gitlab.com/project-emco/core/emco-base/src/orchestrator

go 1.16

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

require (
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.18.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.8.0
	github.com/tidwall/gjson v1.14.0
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/yourbasic/graph v0.0.0-20210606180040-8ecfec1c2869
	gitlab.com/project-emco/core/emco-base/src/clm v0.0.0-00010101000000-000000000000
	gitlab.com/project-emco/core/emco-base/src/monitor v0.0.0-00010101000000-000000000000
	gitlab.com/project-emco/core/emco-base/src/rsync v0.0.0-00010101000000-000000000000
	go.etcd.io/etcd v3.3.25+incompatible
	go.mongodb.org/mongo-driver v1.9.1
	go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo v0.33.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.28.0
	go.opentelemetry.io/contrib/propagators/b3 v1.8.0
	go.opentelemetry.io/otel v1.8.0
	go.opentelemetry.io/otel/exporters/zipkin v1.8.0
	go.opentelemetry.io/otel/sdk v1.8.0
	go.opentelemetry.io/otel/trace v1.8.0
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	google.golang.org/grpc v1.43.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.8.0
	k8s.io/api v0.23.3
	k8s.io/apimachinery v0.23.3
)
