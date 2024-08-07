// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v5.27.0
// source: proto/product.proto

package pb

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

type ProductStatus_ProductStatus int32

const (
	ProductStatus_UNDEFINED ProductStatus_ProductStatus = 0
	ProductStatus_ACTIVE    ProductStatus_ProductStatus = 1
	ProductStatus_DELETED   ProductStatus_ProductStatus = 2
)

// Enum value maps for ProductStatus_ProductStatus.
var (
	ProductStatus_ProductStatus_name = map[int32]string{
		0: "UNDEFINED",
		1: "ACTIVE",
		2: "DELETED",
	}
	ProductStatus_ProductStatus_value = map[string]int32{
		"UNDEFINED": 0,
		"ACTIVE":    1,
		"DELETED":   2,
	}
)

func (x ProductStatus_ProductStatus) Enum() *ProductStatus_ProductStatus {
	p := new(ProductStatus_ProductStatus)
	*p = x
	return p
}

func (x ProductStatus_ProductStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ProductStatus_ProductStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_product_proto_enumTypes[0].Descriptor()
}

func (ProductStatus_ProductStatus) Type() protoreflect.EnumType {
	return &file_proto_product_proto_enumTypes[0]
}

func (x ProductStatus_ProductStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ProductStatus_ProductStatus.Descriptor instead.
func (ProductStatus_ProductStatus) EnumDescriptor() ([]byte, []int) {
	return file_proto_product_proto_rawDescGZIP(), []int{0, 0}
}

type ProductStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ProductStatus) Reset() {
	*x = ProductStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_product_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProductStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProductStatus) ProtoMessage() {}

func (x *ProductStatus) ProtoReflect() protoreflect.Message {
	mi := &file_proto_product_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProductStatus.ProtoReflect.Descriptor instead.
func (*ProductStatus) Descriptor() ([]byte, []int) {
	return file_proto_product_proto_rawDescGZIP(), []int{0}
}

type Product struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id                         string                      `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name                       string                      `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Serial                     string                      `protobuf:"bytes,3,opt,name=serial,proto3" json:"serial,omitempty"`
	Description                string                      `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	Brand                      *Brand                      `protobuf:"bytes,5,opt,name=brand,proto3" json:"brand,omitempty"`
	Status                     ProductStatus_ProductStatus `protobuf:"varint,6,opt,name=status,proto3,enum=proto.ProductStatus_ProductStatus" json:"status,omitempty"`
	ProductCategory            *ProductCategory            `protobuf:"bytes,7,opt,name=product_category,json=productCategory,proto3" json:"product_category,omitempty"`
	ProductSpecs               *ProductSpecs               `protobuf:"bytes,8,opt,name=product_specs,json=productSpecs,proto3" json:"product_specs,omitempty"`
	UnitPriceWithoutVatDisplay string                      `protobuf:"bytes,9,opt,name=unit_price_without_vat_display,json=unitPriceWithoutVatDisplay,proto3" json:"unit_price_without_vat_display,omitempty"`
	UnitPriceWithVatDisplay    string                      `protobuf:"bytes,10,opt,name=unit_price_with_vat_display,json=unitPriceWithVatDisplay,proto3" json:"unit_price_with_vat_display,omitempty"`
	Metadata                   *Metadata                   `protobuf:"bytes,11,opt,name=metadata,proto3" json:"metadata,omitempty"`
}

func (x *Product) Reset() {
	*x = Product{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_product_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Product) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Product) ProtoMessage() {}

func (x *Product) ProtoReflect() protoreflect.Message {
	mi := &file_proto_product_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Product.ProtoReflect.Descriptor instead.
func (*Product) Descriptor() ([]byte, []int) {
	return file_proto_product_proto_rawDescGZIP(), []int{1}
}

func (x *Product) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Product) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Product) GetSerial() string {
	if x != nil {
		return x.Serial
	}
	return ""
}

func (x *Product) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Product) GetBrand() *Brand {
	if x != nil {
		return x.Brand
	}
	return nil
}

func (x *Product) GetStatus() ProductStatus_ProductStatus {
	if x != nil {
		return x.Status
	}
	return ProductStatus_UNDEFINED
}

func (x *Product) GetProductCategory() *ProductCategory {
	if x != nil {
		return x.ProductCategory
	}
	return nil
}

func (x *Product) GetProductSpecs() *ProductSpecs {
	if x != nil {
		return x.ProductSpecs
	}
	return nil
}

func (x *Product) GetUnitPriceWithoutVatDisplay() string {
	if x != nil {
		return x.UnitPriceWithoutVatDisplay
	}
	return ""
}

func (x *Product) GetUnitPriceWithVatDisplay() string {
	if x != nil {
		return x.UnitPriceWithVatDisplay
	}
	return ""
}

func (x *Product) GetMetadata() *Metadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

type ProductsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Length   int64      `protobuf:"varint,1,opt,name=length,proto3" json:"length,omitempty"`
	Products []*Product `protobuf:"bytes,2,rep,name=products,proto3" json:"products,omitempty"`
}

func (x *ProductsResponse) Reset() {
	*x = ProductsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_product_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProductsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProductsResponse) ProtoMessage() {}

func (x *ProductsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_product_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProductsResponse.ProtoReflect.Descriptor instead.
func (*ProductsResponse) Descriptor() ([]byte, []int) {
	return file_proto_product_proto_rawDescGZIP(), []int{2}
}

func (x *ProductsResponse) GetLength() int64 {
	if x != nil {
		return x.Length
	}
	return 0
}

func (x *ProductsResponse) GetProducts() []*Product {
	if x != nil {
		return x.Products
	}
	return nil
}

type ProductStatusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status ProductStatus_ProductStatus `protobuf:"varint,1,opt,name=status,proto3,enum=proto.ProductStatus_ProductStatus" json:"status,omitempty"`
	SortBy *SortBy                     `protobuf:"bytes,2,opt,name=sort_by,json=sortBy,proto3" json:"sort_by,omitempty"`
}

func (x *ProductStatusRequest) Reset() {
	*x = ProductStatusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_product_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProductStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProductStatusRequest) ProtoMessage() {}

func (x *ProductStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_product_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProductStatusRequest.ProtoReflect.Descriptor instead.
func (*ProductStatusRequest) Descriptor() ([]byte, []int) {
	return file_proto_product_proto_rawDescGZIP(), []int{3}
}

func (x *ProductStatusRequest) GetStatus() ProductStatus_ProductStatus {
	if x != nil {
		return x.Status
	}
	return ProductStatus_UNDEFINED
}

func (x *ProductStatusRequest) GetSortBy() *SortBy {
	if x != nil {
		return x.SortBy
	}
	return nil
}

var File_proto_product_proto protoreflect.FileDescriptor

var file_proto_product_proto_rawDesc = []byte{
	0x0a, 0x13, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x11, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x62, 0x72, 0x61, 0x6e, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x12, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x64, 0x75,
	0x63, 0x74, 0x5f, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x19, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74,
	0x5f, 0x73, 0x70, 0x65, 0x63, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x48, 0x0a, 0x0d,
	0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x37, 0x0a,
	0x0d, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0d,
	0x0a, 0x09, 0x55, 0x4e, 0x44, 0x45, 0x46, 0x49, 0x4e, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0a, 0x0a,
	0x06, 0x41, 0x43, 0x54, 0x49, 0x56, 0x45, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x44, 0x45, 0x4c,
	0x45, 0x54, 0x45, 0x44, 0x10, 0x02, 0x22, 0xf3, 0x03, 0x0a, 0x07, 0x50, 0x72, 0x6f, 0x64, 0x75,
	0x63, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x12, 0x20,
	0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x22, 0x0a, 0x05, 0x62, 0x72, 0x61, 0x6e, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x42, 0x72, 0x61, 0x6e, 0x64, 0x52, 0x05, 0x62,
	0x72, 0x61, 0x6e, 0x64, 0x12, 0x3a, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x22, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x50, 0x72, 0x6f,
	0x64, 0x75, 0x63, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x64, 0x75,
	0x63, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x41, 0x0a, 0x10, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f, 0x63, 0x61, 0x74, 0x65,
	0x67, 0x6f, 0x72, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f,
	0x72, 0x79, 0x52, 0x0f, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x43, 0x61, 0x74, 0x65, 0x67,
	0x6f, 0x72, 0x79, 0x12, 0x38, 0x0a, 0x0d, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f, 0x73,
	0x70, 0x65, 0x63, 0x73, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x53, 0x70, 0x65, 0x63, 0x73, 0x52,
	0x0c, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x53, 0x70, 0x65, 0x63, 0x73, 0x12, 0x42, 0x0a,
	0x1e, 0x75, 0x6e, 0x69, 0x74, 0x5f, 0x70, 0x72, 0x69, 0x63, 0x65, 0x5f, 0x77, 0x69, 0x74, 0x68,
	0x6f, 0x75, 0x74, 0x5f, 0x76, 0x61, 0x74, 0x5f, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x18,
	0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x1a, 0x75, 0x6e, 0x69, 0x74, 0x50, 0x72, 0x69, 0x63, 0x65,
	0x57, 0x69, 0x74, 0x68, 0x6f, 0x75, 0x74, 0x56, 0x61, 0x74, 0x44, 0x69, 0x73, 0x70, 0x6c, 0x61,
	0x79, 0x12, 0x3c, 0x0a, 0x1b, 0x75, 0x6e, 0x69, 0x74, 0x5f, 0x70, 0x72, 0x69, 0x63, 0x65, 0x5f,
	0x77, 0x69, 0x74, 0x68, 0x5f, 0x76, 0x61, 0x74, 0x5f, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79,
	0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x17, 0x75, 0x6e, 0x69, 0x74, 0x50, 0x72, 0x69, 0x63,
	0x65, 0x57, 0x69, 0x74, 0x68, 0x56, 0x61, 0x74, 0x44, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x12,
	0x2b, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x0b, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x22, 0x56, 0x0a, 0x10,
	0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x16, 0x0a, 0x06, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x06, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x12, 0x2a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x64,
	0x75, 0x63, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x64,
	0x75, 0x63, 0x74, 0x73, 0x22, 0x7a, 0x0a, 0x14, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x3a, 0x0a, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x22, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x26, 0x0a, 0x07, 0x73, 0x6f, 0x72, 0x74,
	0x5f, 0x62, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x53, 0x6f, 0x72, 0x74, 0x42, 0x79, 0x52, 0x06, 0x73, 0x6f, 0x72, 0x74, 0x42, 0x79,
	0x32, 0x9d, 0x01, 0x0a, 0x0e, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x34, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63,
	0x74, 0x42, 0x79, 0x49, 0x44, 0x12, 0x10, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x49, 0x44,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x22, 0x00, 0x12, 0x55, 0x0a, 0x1b, 0x4c, 0x69, 0x73,
	0x74, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x73, 0x42, 0x79, 0x50, 0x72, 0x6f, 0x64, 0x75,
	0x63, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x50, 0x72,
	0x6f, 0x64, 0x75, 0x63, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x42, 0x0c, 0x5a, 0x0a, 0x63, 0x63, 0x68, 0x6f, 0x69, 0x63, 0x65, 0x2f, 0x70, 0x62, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_product_proto_rawDescOnce sync.Once
	file_proto_product_proto_rawDescData = file_proto_product_proto_rawDesc
)

func file_proto_product_proto_rawDescGZIP() []byte {
	file_proto_product_proto_rawDescOnce.Do(func() {
		file_proto_product_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_product_proto_rawDescData)
	})
	return file_proto_product_proto_rawDescData
}

var file_proto_product_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_product_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_proto_product_proto_goTypes = []interface{}{
	(ProductStatus_ProductStatus)(0), // 0: proto.ProductStatus.ProductStatus
	(*ProductStatus)(nil),            // 1: proto.ProductStatus
	(*Product)(nil),                  // 2: proto.Product
	(*ProductsResponse)(nil),         // 3: proto.ProductsResponse
	(*ProductStatusRequest)(nil),     // 4: proto.ProductStatusRequest
	(*Brand)(nil),                    // 5: proto.Brand
	(*ProductCategory)(nil),          // 6: proto.ProductCategory
	(*ProductSpecs)(nil),             // 7: proto.ProductSpecs
	(*Metadata)(nil),                 // 8: proto.Metadata
	(*SortBy)(nil),                   // 9: proto.SortBy
	(*IDRequest)(nil),                // 10: proto.IDRequest
}
var file_proto_product_proto_depIdxs = []int32{
	5,  // 0: proto.Product.brand:type_name -> proto.Brand
	0,  // 1: proto.Product.status:type_name -> proto.ProductStatus.ProductStatus
	6,  // 2: proto.Product.product_category:type_name -> proto.ProductCategory
	7,  // 3: proto.Product.product_specs:type_name -> proto.ProductSpecs
	8,  // 4: proto.Product.metadata:type_name -> proto.Metadata
	2,  // 5: proto.ProductsResponse.products:type_name -> proto.Product
	0,  // 6: proto.ProductStatusRequest.status:type_name -> proto.ProductStatus.ProductStatus
	9,  // 7: proto.ProductStatusRequest.sort_by:type_name -> proto.SortBy
	10, // 8: proto.ProductService.GetProductByID:input_type -> proto.IDRequest
	4,  // 9: proto.ProductService.ListProductsByProductStatus:input_type -> proto.ProductStatusRequest
	2,  // 10: proto.ProductService.GetProductByID:output_type -> proto.Product
	3,  // 11: proto.ProductService.ListProductsByProductStatus:output_type -> proto.ProductsResponse
	10, // [10:12] is the sub-list for method output_type
	8,  // [8:10] is the sub-list for method input_type
	8,  // [8:8] is the sub-list for extension type_name
	8,  // [8:8] is the sub-list for extension extendee
	0,  // [0:8] is the sub-list for field type_name
}

func init() { file_proto_product_proto_init() }
func file_proto_product_proto_init() {
	if File_proto_product_proto != nil {
		return
	}
	file_proto_brand_proto_init()
	file_proto_common_proto_init()
	file_proto_product_category_proto_init()
	file_proto_product_specs_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_proto_product_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProductStatus); i {
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
		file_proto_product_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Product); i {
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
		file_proto_product_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProductsResponse); i {
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
		file_proto_product_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProductStatusRequest); i {
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
			RawDescriptor: file_proto_product_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_product_proto_goTypes,
		DependencyIndexes: file_proto_product_proto_depIdxs,
		EnumInfos:         file_proto_product_proto_enumTypes,
		MessageInfos:      file_proto_product_proto_msgTypes,
	}.Build()
	File_proto_product_proto = out.File
	file_proto_product_proto_rawDesc = nil
	file_proto_product_proto_goTypes = nil
	file_proto_product_proto_depIdxs = nil
}
