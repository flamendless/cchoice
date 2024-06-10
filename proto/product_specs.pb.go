// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v5.27.0
// source: proto/product_specs.proto

package pb

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ProductSpecs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ID            int64  `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Colours       string `protobuf:"bytes,2,opt,name=colours,proto3" json:"colours,omitempty"`
	Sizes         string `protobuf:"bytes,3,opt,name=sizes,proto3" json:"sizes,omitempty"`
	Segmentation  string `protobuf:"bytes,4,opt,name=segmentation,proto3" json:"segmentation,omitempty"`
	PartNumber    string `protobuf:"bytes,5,opt,name=part_number,json=partNumber,proto3" json:"part_number,omitempty"`
	Power         string `protobuf:"bytes,6,opt,name=power,proto3" json:"power,omitempty"`
	Capacity      string `protobuf:"bytes,7,opt,name=capacity,proto3" json:"capacity,omitempty"`
	ScopeOfSupply string `protobuf:"bytes,8,opt,name=scope_of_supply,json=scopeOfSupply,proto3" json:"scope_of_supply,omitempty"`
}

func (x *ProductSpecs) Reset() {
	*x = ProductSpecs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_product_specs_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProductSpecs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProductSpecs) ProtoMessage() {}

func (x *ProductSpecs) ProtoReflect() protoreflect.Message {
	mi := &file_proto_product_specs_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProductSpecs.ProtoReflect.Descriptor instead.
func (*ProductSpecs) Descriptor() ([]byte, []int) {
	return file_proto_product_specs_proto_rawDescGZIP(), []int{0}
}

func (x *ProductSpecs) GetID() int64 {
	if x != nil {
		return x.ID
	}
	return 0
}

func (x *ProductSpecs) GetColours() string {
	if x != nil {
		return x.Colours
	}
	return ""
}

func (x *ProductSpecs) GetSizes() string {
	if x != nil {
		return x.Sizes
	}
	return ""
}

func (x *ProductSpecs) GetSegmentation() string {
	if x != nil {
		return x.Segmentation
	}
	return ""
}

func (x *ProductSpecs) GetPartNumber() string {
	if x != nil {
		return x.PartNumber
	}
	return ""
}

func (x *ProductSpecs) GetPower() string {
	if x != nil {
		return x.Power
	}
	return ""
}

func (x *ProductSpecs) GetCapacity() string {
	if x != nil {
		return x.Capacity
	}
	return ""
}

func (x *ProductSpecs) GetScopeOfSupply() string {
	if x != nil {
		return x.ScopeOfSupply
	}
	return ""
}

var File_proto_product_specs_proto protoreflect.FileDescriptor

var file_proto_product_specs_proto_rawDesc = []byte{
	0x0a, 0x19, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f,
	0x73, 0x70, 0x65, 0x63, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0xed, 0x01, 0x0a, 0x0c, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x53, 0x70,
	0x65, 0x63, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x02, 0x49, 0x44, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6c, 0x6f, 0x75, 0x72, 0x73, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6c, 0x6f, 0x75, 0x72, 0x73, 0x12, 0x14, 0x0a,
	0x05, 0x73, 0x69, 0x7a, 0x65, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x73, 0x69,
	0x7a, 0x65, 0x73, 0x12, 0x22, 0x0a, 0x0c, 0x73, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x73, 0x65, 0x67, 0x6d, 0x65,
	0x6e, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1f, 0x0a, 0x0b, 0x70, 0x61, 0x72, 0x74, 0x5f,
	0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x61,
	0x72, 0x74, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x6f, 0x77, 0x65,
	0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x70, 0x6f, 0x77, 0x65, 0x72, 0x12, 0x1a,
	0x0a, 0x08, 0x63, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x63, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x12, 0x26, 0x0a, 0x0f, 0x73, 0x63,
	0x6f, 0x70, 0x65, 0x5f, 0x6f, 0x66, 0x5f, 0x73, 0x75, 0x70, 0x70, 0x6c, 0x79, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0d, 0x73, 0x63, 0x6f, 0x70, 0x65, 0x4f, 0x66, 0x53, 0x75, 0x70, 0x70,
	0x6c, 0x79, 0x42, 0x0c, 0x5a, 0x0a, 0x63, 0x63, 0x68, 0x6f, 0x69, 0x63, 0x65, 0x2f, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_product_specs_proto_rawDescOnce sync.Once
	file_proto_product_specs_proto_rawDescData = file_proto_product_specs_proto_rawDesc
)

func file_proto_product_specs_proto_rawDescGZIP() []byte {
	file_proto_product_specs_proto_rawDescOnce.Do(func() {
		file_proto_product_specs_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_product_specs_proto_rawDescData)
	})
	return file_proto_product_specs_proto_rawDescData
}

var file_proto_product_specs_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_proto_product_specs_proto_goTypes = []interface{}{
	(*ProductSpecs)(nil), // 0: proto.ProductSpecs
}
var file_proto_product_specs_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_product_specs_proto_init() }
func file_proto_product_specs_proto_init() {
	if File_proto_product_specs_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_product_specs_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProductSpecs); i {
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
			RawDescriptor: file_proto_product_specs_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_product_specs_proto_goTypes,
		DependencyIndexes: file_proto_product_specs_proto_depIdxs,
		MessageInfos:      file_proto_product_specs_proto_msgTypes,
	}.Build()
	File_proto_product_specs_proto = out.File
	file_proto_product_specs_proto_rawDesc = nil
	file_proto_product_specs_proto_goTypes = nil
	file_proto_product_specs_proto_depIdxs = nil
}
