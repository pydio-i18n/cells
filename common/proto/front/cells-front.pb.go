// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: cells-front.proto

package front

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

type ExposedParametersRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Scope   string `protobuf:"bytes,1,opt,name=Scope,proto3" json:"Scope,omitempty"`
	Exposed bool   `protobuf:"varint,2,opt,name=Exposed,proto3" json:"Exposed,omitempty"`
}

func (x *ExposedParametersRequest) Reset() {
	*x = ExposedParametersRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cells_front_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExposedParametersRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExposedParametersRequest) ProtoMessage() {}

func (x *ExposedParametersRequest) ProtoReflect() protoreflect.Message {
	mi := &file_cells_front_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExposedParametersRequest.ProtoReflect.Descriptor instead.
func (*ExposedParametersRequest) Descriptor() ([]byte, []int) {
	return file_cells_front_proto_rawDescGZIP(), []int{0}
}

func (x *ExposedParametersRequest) GetScope() string {
	if x != nil {
		return x.Scope
	}
	return ""
}

func (x *ExposedParametersRequest) GetExposed() bool {
	if x != nil {
		return x.Exposed
	}
	return false
}

type ExposedParameter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name     string `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	Scope    string `protobuf:"bytes,2,opt,name=Scope,proto3" json:"Scope,omitempty"`
	PluginId string `protobuf:"bytes,3,opt,name=PluginId,proto3" json:"PluginId,omitempty"`
}

func (x *ExposedParameter) Reset() {
	*x = ExposedParameter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cells_front_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExposedParameter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExposedParameter) ProtoMessage() {}

func (x *ExposedParameter) ProtoReflect() protoreflect.Message {
	mi := &file_cells_front_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExposedParameter.ProtoReflect.Descriptor instead.
func (*ExposedParameter) Descriptor() ([]byte, []int) {
	return file_cells_front_proto_rawDescGZIP(), []int{1}
}

func (x *ExposedParameter) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ExposedParameter) GetScope() string {
	if x != nil {
		return x.Scope
	}
	return ""
}

func (x *ExposedParameter) GetPluginId() string {
	if x != nil {
		return x.PluginId
	}
	return ""
}

type ExposedParametersResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Parameters []*ExposedParameter `protobuf:"bytes,1,rep,name=Parameters,proto3" json:"Parameters,omitempty"`
}

func (x *ExposedParametersResponse) Reset() {
	*x = ExposedParametersResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cells_front_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExposedParametersResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExposedParametersResponse) ProtoMessage() {}

func (x *ExposedParametersResponse) ProtoReflect() protoreflect.Message {
	mi := &file_cells_front_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExposedParametersResponse.ProtoReflect.Descriptor instead.
func (*ExposedParametersResponse) Descriptor() ([]byte, []int) {
	return file_cells_front_proto_rawDescGZIP(), []int{2}
}

func (x *ExposedParametersResponse) GetParameters() []*ExposedParameter {
	if x != nil {
		return x.Parameters
	}
	return nil
}

var File_cells_front_proto protoreflect.FileDescriptor

var file_cells_front_proto_rawDesc = []byte{
	0x0a, 0x11, 0x63, 0x65, 0x6c, 0x6c, 0x73, 0x2d, 0x66, 0x72, 0x6f, 0x6e, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x05, 0x66, 0x72, 0x6f, 0x6e, 0x74, 0x22, 0x4a, 0x0a, 0x18, 0x45, 0x78,
	0x70, 0x6f, 0x73, 0x65, 0x64, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x53, 0x63, 0x6f, 0x70, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x53, 0x63, 0x6f, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x45, 0x78, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x45,
	0x78, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x22, 0x58, 0x0a, 0x10, 0x45, 0x78, 0x70, 0x6f, 0x73, 0x65,
	0x64, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x4e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x14,
	0x0a, 0x05, 0x53, 0x63, 0x6f, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x53,
	0x63, 0x6f, 0x70, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x49, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x49, 0x64,
	0x22, 0x54, 0x0a, 0x19, 0x45, 0x78, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x50, 0x61, 0x72, 0x61, 0x6d,
	0x65, 0x74, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x37, 0x0a,
	0x0a, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x17, 0x2e, 0x66, 0x72, 0x6f, 0x6e, 0x74, 0x2e, 0x45, 0x78, 0x70, 0x6f, 0x73, 0x65,
	0x64, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x52, 0x0a, 0x50, 0x61, 0x72, 0x61,
	0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x32, 0x6b, 0x0a, 0x0f, 0x4d, 0x61, 0x6e, 0x69, 0x66, 0x65,
	0x73, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x58, 0x0a, 0x11, 0x45, 0x78, 0x70,
	0x6f, 0x73, 0x65, 0x64, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x12, 0x1f,
	0x2e, 0x66, 0x72, 0x6f, 0x6e, 0x74, 0x2e, 0x45, 0x78, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x50, 0x61,
	0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x20, 0x2e, 0x66, 0x72, 0x6f, 0x6e, 0x74, 0x2e, 0x45, 0x78, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x50,
	0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x42, 0x2e, 0x5a, 0x2c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x70, 0x79, 0x64, 0x69, 0x6f, 0x2f, 0x63, 0x65, 0x6c, 0x6c, 0x73, 0x2f, 0x76, 0x35,
	0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x66, 0x72,
	0x6f, 0x6e, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cells_front_proto_rawDescOnce sync.Once
	file_cells_front_proto_rawDescData = file_cells_front_proto_rawDesc
)

func file_cells_front_proto_rawDescGZIP() []byte {
	file_cells_front_proto_rawDescOnce.Do(func() {
		file_cells_front_proto_rawDescData = protoimpl.X.CompressGZIP(file_cells_front_proto_rawDescData)
	})
	return file_cells_front_proto_rawDescData
}

var file_cells_front_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_cells_front_proto_goTypes = []any{
	(*ExposedParametersRequest)(nil),  // 0: front.ExposedParametersRequest
	(*ExposedParameter)(nil),          // 1: front.ExposedParameter
	(*ExposedParametersResponse)(nil), // 2: front.ExposedParametersResponse
}
var file_cells_front_proto_depIdxs = []int32{
	1, // 0: front.ExposedParametersResponse.Parameters:type_name -> front.ExposedParameter
	0, // 1: front.ManifestService.ExposedParameters:input_type -> front.ExposedParametersRequest
	2, // 2: front.ManifestService.ExposedParameters:output_type -> front.ExposedParametersResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_cells_front_proto_init() }
func file_cells_front_proto_init() {
	if File_cells_front_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cells_front_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*ExposedParametersRequest); i {
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
		file_cells_front_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*ExposedParameter); i {
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
		file_cells_front_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*ExposedParametersResponse); i {
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
			RawDescriptor: file_cells_front_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_cells_front_proto_goTypes,
		DependencyIndexes: file_cells_front_proto_depIdxs,
		MessageInfos:      file_cells_front_proto_msgTypes,
	}.Build()
	File_cells_front_proto = out.File
	file_cells_front_proto_rawDesc = nil
	file_cells_front_proto_goTypes = nil
	file_cells_front_proto_depIdxs = nil
}
