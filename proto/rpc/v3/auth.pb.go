// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: proto/rpc/v3/auth.proto

package authv3

import (
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_proto_rpc_v3_auth_proto protoreflect.FileDescriptor

var file_proto_rpc_v3_auth_proto_rawDesc = []byte{
	0x0a, 0x17, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x72, 0x70, 0x63, 0x2f, 0x76, 0x33, 0x2f, 0x61,
	0x75, 0x74, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x17, 0x70, 0x61, 0x72, 0x61, 0x6c,
	0x75, 0x73, 0x2e, 0x64, 0x65, 0x76, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e,
	0x76, 0x33, 0x1a, 0x22, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x70, 0x62, 0x2f, 0x76, 0x33, 0x2f, 0x61, 0x75, 0x74, 0x68,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0x8e, 0x01, 0x0a, 0x0b, 0x41, 0x75, 0x74, 0x68, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x7f, 0x0a, 0x10, 0x49, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x41, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x12, 0x34, 0x2e, 0x70, 0x61, 0x72,
	0x61, 0x6c, 0x75, 0x73, 0x2e, 0x64, 0x65, 0x76, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x33, 0x2e, 0x49, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x41, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x35, 0x2e, 0x70, 0x61, 0x72, 0x61, 0x6c, 0x75, 0x73, 0x2e, 0x64, 0x65, 0x76, 0x2e, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x33, 0x2e, 0x49,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x41, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0xd9, 0x01, 0x0a, 0x1b, 0x63, 0x6f, 0x6d, 0x2e,
	0x70, 0x61, 0x72, 0x61, 0x6c, 0x75, 0x73, 0x2e, 0x64, 0x65, 0x76, 0x2e, 0x72, 0x70, 0x63, 0x2e,
	0x61, 0x75, 0x74, 0x68, 0x2e, 0x76, 0x33, 0x42, 0x09, 0x41, 0x75, 0x74, 0x68, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x70, 0x61, 0x72, 0x61, 0x6c, 0x75, 0x73, 0x2f, 0x70, 0x61, 0x72, 0x61, 0x6c, 0x75, 0x73,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x72, 0x70, 0x63, 0x2f, 0x76, 0x33, 0x3b, 0x61, 0x75,
	0x74, 0x68, 0x76, 0x33, 0xa2, 0x02, 0x04, 0x50, 0x44, 0x52, 0x41, 0xaa, 0x02, 0x17, 0x50, 0x61,
	0x72, 0x61, 0x6c, 0x75, 0x73, 0x2e, 0x44, 0x65, 0x76, 0x2e, 0x52, 0x70, 0x63, 0x2e, 0x41, 0x75,
	0x74, 0x68, 0x2e, 0x56, 0x33, 0xca, 0x02, 0x17, 0x50, 0x61, 0x72, 0x61, 0x6c, 0x75, 0x73, 0x5c,
	0x44, 0x65, 0x76, 0x5c, 0x52, 0x70, 0x63, 0x5c, 0x41, 0x75, 0x74, 0x68, 0x5c, 0x56, 0x33, 0xe2,
	0x02, 0x23, 0x50, 0x61, 0x72, 0x61, 0x6c, 0x75, 0x73, 0x5c, 0x44, 0x65, 0x76, 0x5c, 0x52, 0x70,
	0x63, 0x5c, 0x41, 0x75, 0x74, 0x68, 0x5c, 0x56, 0x33, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x1b, 0x50, 0x61, 0x72, 0x61, 0x6c, 0x75, 0x73, 0x3a,
	0x3a, 0x44, 0x65, 0x76, 0x3a, 0x3a, 0x52, 0x70, 0x63, 0x3a, 0x3a, 0x41, 0x75, 0x74, 0x68, 0x3a,
	0x3a, 0x56, 0x33, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_proto_rpc_v3_auth_proto_goTypes = []interface{}{
	(*v3.IsRequestAllowedRequest)(nil),  // 0: paralus.dev.types.common.v3.IsRequestAllowedRequest
	(*v3.IsRequestAllowedResponse)(nil), // 1: paralus.dev.types.common.v3.IsRequestAllowedResponse
}
var file_proto_rpc_v3_auth_proto_depIdxs = []int32{
	0, // 0: paralus.dev.rpc.auth.v3.AuthService.IsRequestAllowed:input_type -> paralus.dev.types.common.v3.IsRequestAllowedRequest
	1, // 1: paralus.dev.rpc.auth.v3.AuthService.IsRequestAllowed:output_type -> paralus.dev.types.common.v3.IsRequestAllowedResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_rpc_v3_auth_proto_init() }
func file_proto_rpc_v3_auth_proto_init() {
	if File_proto_rpc_v3_auth_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_rpc_v3_auth_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_rpc_v3_auth_proto_goTypes,
		DependencyIndexes: file_proto_rpc_v3_auth_proto_depIdxs,
	}.Build()
	File_proto_rpc_v3_auth_proto = out.File
	file_proto_rpc_v3_auth_proto_rawDesc = nil
	file_proto_rpc_v3_auth_proto_goTypes = nil
	file_proto_rpc_v3_auth_proto_depIdxs = nil
}