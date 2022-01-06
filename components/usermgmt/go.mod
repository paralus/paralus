module github.com/RafaySystems/rcloud-base/components/usermgmt

go 1.16

require (
	github.com/RafaySystems/rcloud-base/components/adminsrv v0.0.0-unpublished
	github.com/RafaySystems/rcloud-base/components/common v0.0.0-unpublished
	github.com/gogo/protobuf v1.3.2
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.2
	github.com/ory/kratos-client-go v0.8.2-alpha.1
	github.com/spf13/viper v1.10.1
	github.com/uptrace/bun v1.0.20
	github.com/uptrace/bun/dialect/pgdialect v1.0.20
	github.com/uptrace/bun/driver/pgdriver v1.0.20
	github.com/uptrace/bun/extra/bundebug v1.0.20
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa
	google.golang.org/grpc v1.43.0
	google.golang.org/protobuf v1.27.1
	sigs.k8s.io/controller-runtime v0.11.0
)

replace (
	github.com/RafaySystems/rcloud-base/components/adminsrv v0.0.0-unpublished => ../adminsrv/
	github.com/RafaySystems/rcloud-base/components/common v0.0.0-unpublished => ../common/
)
