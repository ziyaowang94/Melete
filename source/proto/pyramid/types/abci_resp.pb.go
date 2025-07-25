// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v5.26.1
// source: proto/pyramid/types/abci_resp.proto

package types

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ABCIExecutionReceipt struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code    int32             `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Log     string            `protobuf:"bytes,2,opt,name=log,proto3" json:"log,omitempty"`
	Info    string            `protobuf:"bytes,3,opt,name=info,proto3" json:"info,omitempty"`
	Options map[string]string `protobuf:"bytes,4,rep,name=options,proto3" json:"options,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *ABCIExecutionReceipt) Reset() {
	*x = ABCIExecutionReceipt{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_pyramid_types_abci_resp_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ABCIExecutionReceipt) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ABCIExecutionReceipt) ProtoMessage() {}

func (x *ABCIExecutionReceipt) ProtoReflect() protoreflect.Message {
	mi := &file_proto_pyramid_types_abci_resp_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ABCIExecutionReceipt.ProtoReflect.Descriptor instead.
func (*ABCIExecutionReceipt) Descriptor() ([]byte, []int) {
	return file_proto_pyramid_types_abci_resp_proto_rawDescGZIP(), []int{0}
}

func (x *ABCIExecutionReceipt) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *ABCIExecutionReceipt) GetLog() string {
	if x != nil {
		return x.Log
	}
	return ""
}

func (x *ABCIExecutionReceipt) GetInfo() string {
	if x != nil {
		return x.Info
	}
	return ""
}

func (x *ABCIExecutionReceipt) GetOptions() map[string]string {
	if x != nil {
		return x.Options
	}
	return nil
}

var File_proto_pyramid_types_abci_resp_proto protoreflect.FileDescriptor

var file_proto_pyramid_types_abci_resp_proto_rawDesc = []byte{
	0x0a, 0x23, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x79, 0x72, 0x61, 0x6d, 0x69, 0x64, 0x2f,
	0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x61, 0x62, 0x63, 0x69, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x70, 0x79, 0x72, 0x61, 0x6d, 0x69, 0x64, 0x2e, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x22, 0xd8, 0x01, 0x0a, 0x14, 0x41, 0x42, 0x43, 0x49, 0x45, 0x78, 0x65,
	0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x12, 0x12, 0x0a,
	0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x63, 0x6f, 0x64,
	0x65, 0x12, 0x10, 0x0a, 0x03, 0x6c, 0x6f, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6c, 0x6f, 0x67, 0x12, 0x12, 0x0a, 0x04, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x69, 0x6e, 0x66, 0x6f, 0x12, 0x4a, 0x0a, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x30, 0x2e, 0x70, 0x79, 0x72, 0x61, 0x6d,
	0x69, 0x64, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x41, 0x42, 0x43, 0x49, 0x45, 0x78, 0x65,
	0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x2e, 0x4f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x6f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x1a, 0x3a, 0x0a, 0x0c, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42,
	0x1e, 0x5a, 0x1c, 0x65, 0x6d, 0x75, 0x6c, 0x61, 0x74, 0x6f, 0x72, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2f, 0x70, 0x79, 0x72, 0x61, 0x6d, 0x69, 0x64, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_pyramid_types_abci_resp_proto_rawDescOnce sync.Once
	file_proto_pyramid_types_abci_resp_proto_rawDescData = file_proto_pyramid_types_abci_resp_proto_rawDesc
)

func file_proto_pyramid_types_abci_resp_proto_rawDescGZIP() []byte {
	file_proto_pyramid_types_abci_resp_proto_rawDescOnce.Do(func() {
		file_proto_pyramid_types_abci_resp_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_pyramid_types_abci_resp_proto_rawDescData)
	})
	return file_proto_pyramid_types_abci_resp_proto_rawDescData
}

var file_proto_pyramid_types_abci_resp_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_pyramid_types_abci_resp_proto_goTypes = []interface{}{
	(*ABCIExecutionReceipt)(nil), // 0: pyramid.types.ABCIExecutionReceipt
	nil,                          // 1: pyramid.types.ABCIExecutionReceipt.OptionsEntry
}
var file_proto_pyramid_types_abci_resp_proto_depIdxs = []int32{
	1, // 0: pyramid.types.ABCIExecutionReceipt.options:type_name -> pyramid.types.ABCIExecutionReceipt.OptionsEntry
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_proto_pyramid_types_abci_resp_proto_init() }
func file_proto_pyramid_types_abci_resp_proto_init() {
	if File_proto_pyramid_types_abci_resp_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_pyramid_types_abci_resp_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ABCIExecutionReceipt); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_pyramid_types_abci_resp_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_pyramid_types_abci_resp_proto_goTypes,
		DependencyIndexes: file_proto_pyramid_types_abci_resp_proto_depIdxs,
		MessageInfos:      file_proto_pyramid_types_abci_resp_proto_msgTypes,
	}.Build()
	File_proto_pyramid_types_abci_resp_proto = out.File
	file_proto_pyramid_types_abci_resp_proto_rawDesc = nil
	file_proto_pyramid_types_abci_resp_proto_goTypes = nil
	file_proto_pyramid_types_abci_resp_proto_depIdxs = nil
}
