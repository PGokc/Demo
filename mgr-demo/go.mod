module mgr-demo

go 1.20

require (
	code.byted.org/infcs/lib-log v0.0.11
	code.byted.org/infcs/mgr v1.0.0
)

require (
	code.byted.org/gopkg/consul v1.2.4 // indirect
	code.byted.org/gopkg/metrics v1.4.25 // indirect
	code.byted.org/kite/kitex v1.15.1 // indirect
	github.com/apache/thrift v0.17.0 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/jonboulle/clockwork v0.3.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20200724154423-2164a8ac840e // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

require (
	code.byted.org/aiops/apm_vendor_byted v0.0.22 // indirect
	code.byted.org/aiops/metrics_codec v0.0.18 // indirect
	code.byted.org/aiops/monitoring-common-go v0.0.3 // indirect
	code.byted.org/gopkg/apm_vendor_interface v0.0.2 // indirect
	code.byted.org/gopkg/ctxvalues v0.4.0 // indirect
	code.byted.org/gopkg/env v1.5.8 // indirect
	code.byted.org/gopkg/logs v1.2.12 // indirect
	code.byted.org/gopkg/metrics_core v0.0.26 // indirect
	code.byted.org/gopkg/net2 v1.5.0 // indirect
	code.byted.org/kite/endpoint v3.7.5+incompatible // indirect
	code.byted.org/log_market/gosdk v0.0.0-20220328031951-809cbf0ba485 // indirect
	code.byted.org/security/sensitive_finder_engine v0.3.16 // indirect
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/gopkg v0.1.1 // indirect
	github.com/caarlos0/env/v6 v6.2.2 // indirect
	github.com/choleraehyq/pid v0.0.21 // indirect
	github.com/cloudwego/frugal v0.1.1 // indirect
	github.com/cloudwego/kitex v0.3.4
	github.com/cloudwego/netpoll v0.2.5 // indirect
	github.com/cloudwego/thriftgo v0.1.3 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/gopherjs/gopherjs v1.12.80 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible // indirect
	github.com/lestrrat-go/strftime v1.0.6 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/oleiade/lane v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/tidwall/gjson v1.14.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	golang.org/x/arch v0.11.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/time v0.6.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240314234333-6e1732d8331c // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace git.byted.org/ee/gopkg => git.byted.org/ee/gopkg v1.6.36

replace code.byted.org/rocketmq/rocketmq-go-proxy => code.byted.org/rocketmq/rocketmq-go-proxy v1.4.18

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

replace code.byted.org/kite/kitex => code.byted.org/kite/kitex v1.6.6

replace gorm.io/driver/mysql => gorm.io/driver/mysql v1.0.4

replace gorm.io/gorm => gorm.io/gorm v1.21.10

replace github.com/apache/thrift => github.com/apache/thrift v0.13.0

replace code.byted.org/infcs/mgr => code.byted.org/larkarch/mgr v0.5.8

replace code.byted.org/infcs/rds-lib => code.byted.org/larkarch/rds-lib v1.0.22

replace code.byted.org/infcs/rds-operator => code.byted.org/larkarch/rds-operator v1.0.3

// replace gorm.io/gorm => gorm.io/gorm v1.20.8

// replace code.byted.org/gopkg/gorm/v2 => code.byted.org/gopkg/gorm/v2 v2.0.3

//replace code.byted.org/infcs/rds-lib => /Users/bytedance/go/src/code.byted.org/infcs/rds-lib
//replace code.byted.org/infcs/lib-mgr-common => /Users/bytedance/code/lib-mgr-common

replace code.byted.org/middleware/framework_version_collector_client => code.byted.org/middleware/framework_version_collector_client v1.3.0

replace (
	code.byted.org/beops/beops => code.byted.org/beops/beops v0.0.3-forops
	code.byted.org/beops/biz/approval => code.byted.org/beops/biz/approval v0.2.0-pre10-forops
	code.byted.org/beops/biz/approval_open/sdk => code.byted.org/beops/biz/approval_open/sdk v0.0.1-forops
	code.byted.org/beops/biz/approval_open/types => code.byted.org/beops/biz/approval_open/types v0.0.1-forops
	code.byted.org/beops/biz/user => code.byted.org/beops/biz/user v0.0.13-forops
	github.com/cloudwego/frugal => github.com/cloudwego/frugal v0.1.16
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.5.5
	k8s.io/api => k8s.io/api v0.27.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.27.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.27.2
	k8s.io/apiserver => k8s.io/apiserver v0.27.2
	k8s.io/client-go => k8s.io/client-go v0.27.2
	k8s.io/component-base => k8s.io/component-base v0.27.2
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20230501164219-8b0f38b5fd1f
	k8s.io/utils => k8s.io/utils v0.0.0-20230209194617-a36077c30491
)
