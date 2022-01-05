module github.com/RafaySystems/rcloud-base/components/usermgmt

go 1.16

require (
	github.com/RafaySystems/rcloud-base/components/adminsrv v0.0.0-unpublished
	github.com/RafaySystems/rcloud-base/components/common v0.0.0-unpublished
	github.com/gogo/protobuf v1.3.2
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.2
	github.com/ory/kratos-client-go v0.8.2-alpha.1
	github.com/spf13/viper v1.10.1
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa
	google.golang.org/grpc v1.43.0
	google.golang.org/protobuf v1.27.1
	sigs.k8s.io/controller-runtime v0.11.0
)

replace (
	github.com/RafaySystems/rcloud-base/components/adminsrv v0.0.0-unpublished => ../adminsrv/
	github.com/RafaySystems/rcloud-base/components/common v0.0.0-unpublished => ../common/
)
