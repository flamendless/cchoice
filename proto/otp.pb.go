// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v5.27.0
// source: proto/otp.proto

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

type EnrollOTPRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserId      string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Issuer      string `protobuf:"bytes,2,opt,name=issuer,proto3" json:"issuer,omitempty"`
	AccountName string `protobuf:"bytes,3,opt,name=account_name,json=accountName,proto3" json:"account_name,omitempty"`
}

func (x *EnrollOTPRequest) Reset() {
	*x = EnrollOTPRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_otp_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EnrollOTPRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EnrollOTPRequest) ProtoMessage() {}

func (x *EnrollOTPRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_otp_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EnrollOTPRequest.ProtoReflect.Descriptor instead.
func (*EnrollOTPRequest) Descriptor() ([]byte, []int) {
	return file_proto_otp_proto_rawDescGZIP(), []int{0}
}

func (x *EnrollOTPRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *EnrollOTPRequest) GetIssuer() string {
	if x != nil {
		return x.Issuer
	}
	return ""
}

func (x *EnrollOTPRequest) GetAccountName() string {
	if x != nil {
		return x.AccountName
	}
	return ""
}

type EnrollOTPResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Secret        string `protobuf:"bytes,1,opt,name=secret,proto3" json:"secret,omitempty"`
	RecoveryCodes string `protobuf:"bytes,2,opt,name=recovery_codes,json=recoveryCodes,proto3" json:"recovery_codes,omitempty"`
	Image         []byte `protobuf:"bytes,3,opt,name=image,proto3" json:"image,omitempty"`
}

func (x *EnrollOTPResponse) Reset() {
	*x = EnrollOTPResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_otp_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EnrollOTPResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EnrollOTPResponse) ProtoMessage() {}

func (x *EnrollOTPResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_otp_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EnrollOTPResponse.ProtoReflect.Descriptor instead.
func (*EnrollOTPResponse) Descriptor() ([]byte, []int) {
	return file_proto_otp_proto_rawDescGZIP(), []int{1}
}

func (x *EnrollOTPResponse) GetSecret() string {
	if x != nil {
		return x.Secret
	}
	return ""
}

func (x *EnrollOTPResponse) GetRecoveryCodes() string {
	if x != nil {
		return x.RecoveryCodes
	}
	return ""
}

func (x *EnrollOTPResponse) GetImage() []byte {
	if x != nil {
		return x.Image
	}
	return nil
}

type GenerateOTPCodeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Method OTPMethod_OTPMethod `protobuf:"varint,1,opt,name=method,proto3,enum=proto.OTPMethod_OTPMethod" json:"method,omitempty"`
	UserId string              `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

func (x *GenerateOTPCodeRequest) Reset() {
	*x = GenerateOTPCodeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_otp_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GenerateOTPCodeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenerateOTPCodeRequest) ProtoMessage() {}

func (x *GenerateOTPCodeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_otp_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GenerateOTPCodeRequest.ProtoReflect.Descriptor instead.
func (*GenerateOTPCodeRequest) Descriptor() ([]byte, []int) {
	return file_proto_otp_proto_rawDescGZIP(), []int{2}
}

func (x *GenerateOTPCodeRequest) GetMethod() OTPMethod_OTPMethod {
	if x != nil {
		return x.Method
	}
	return OTPMethod_UNDEFINED
}

func (x *GenerateOTPCodeRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type GenerateOTPCodeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GenerateOTPCodeResponse) Reset() {
	*x = GenerateOTPCodeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_otp_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GenerateOTPCodeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenerateOTPCodeResponse) ProtoMessage() {}

func (x *GenerateOTPCodeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_otp_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GenerateOTPCodeResponse.ProtoReflect.Descriptor instead.
func (*GenerateOTPCodeResponse) Descriptor() ([]byte, []int) {
	return file_proto_otp_proto_rawDescGZIP(), []int{3}
}

type FinishOTPEnrollmentRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserId   string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Passcode string `protobuf:"bytes,2,opt,name=passcode,proto3" json:"passcode,omitempty"`
}

func (x *FinishOTPEnrollmentRequest) Reset() {
	*x = FinishOTPEnrollmentRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_otp_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FinishOTPEnrollmentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FinishOTPEnrollmentRequest) ProtoMessage() {}

func (x *FinishOTPEnrollmentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_otp_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FinishOTPEnrollmentRequest.ProtoReflect.Descriptor instead.
func (*FinishOTPEnrollmentRequest) Descriptor() ([]byte, []int) {
	return file_proto_otp_proto_rawDescGZIP(), []int{4}
}

func (x *FinishOTPEnrollmentRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *FinishOTPEnrollmentRequest) GetPasscode() string {
	if x != nil {
		return x.Passcode
	}
	return ""
}

type FinishOTPEnrollmentResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *FinishOTPEnrollmentResponse) Reset() {
	*x = FinishOTPEnrollmentResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_otp_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FinishOTPEnrollmentResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FinishOTPEnrollmentResponse) ProtoMessage() {}

func (x *FinishOTPEnrollmentResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_otp_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FinishOTPEnrollmentResponse.ProtoReflect.Descriptor instead.
func (*FinishOTPEnrollmentResponse) Descriptor() ([]byte, []int) {
	return file_proto_otp_proto_rawDescGZIP(), []int{5}
}

type GetOTPInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserId    string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	OtpMethod string `protobuf:"bytes,2,opt,name=otp_method,json=otpMethod,proto3" json:"otp_method,omitempty"`
}

func (x *GetOTPInfoRequest) Reset() {
	*x = GetOTPInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_otp_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetOTPInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetOTPInfoRequest) ProtoMessage() {}

func (x *GetOTPInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_otp_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetOTPInfoRequest.ProtoReflect.Descriptor instead.
func (*GetOTPInfoRequest) Descriptor() ([]byte, []int) {
	return file_proto_otp_proto_rawDescGZIP(), []int{6}
}

func (x *GetOTPInfoRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *GetOTPInfoRequest) GetOtpMethod() string {
	if x != nil {
		return x.OtpMethod
	}
	return ""
}

type GetOTPInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Recipient string `protobuf:"bytes,1,opt,name=recipient,proto3" json:"recipient,omitempty"`
}

func (x *GetOTPInfoResponse) Reset() {
	*x = GetOTPInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_otp_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetOTPInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetOTPInfoResponse) ProtoMessage() {}

func (x *GetOTPInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_otp_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetOTPInfoResponse.ProtoReflect.Descriptor instead.
func (*GetOTPInfoResponse) Descriptor() ([]byte, []int) {
	return file_proto_otp_proto_rawDescGZIP(), []int{7}
}

func (x *GetOTPInfoResponse) GetRecipient() string {
	if x != nil {
		return x.Recipient
	}
	return ""
}

type ValidateOTPRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Passcode string `protobuf:"bytes,1,opt,name=passcode,proto3" json:"passcode,omitempty"`
	UserId   string `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

func (x *ValidateOTPRequest) Reset() {
	*x = ValidateOTPRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_otp_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidateOTPRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateOTPRequest) ProtoMessage() {}

func (x *ValidateOTPRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_otp_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateOTPRequest.ProtoReflect.Descriptor instead.
func (*ValidateOTPRequest) Descriptor() ([]byte, []int) {
	return file_proto_otp_proto_rawDescGZIP(), []int{8}
}

func (x *ValidateOTPRequest) GetPasscode() string {
	if x != nil {
		return x.Passcode
	}
	return ""
}

func (x *ValidateOTPRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type ValidateOTPResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Valid bool `protobuf:"varint,1,opt,name=valid,proto3" json:"valid,omitempty"`
}

func (x *ValidateOTPResponse) Reset() {
	*x = ValidateOTPResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_otp_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidateOTPResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateOTPResponse) ProtoMessage() {}

func (x *ValidateOTPResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_otp_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateOTPResponse.ProtoReflect.Descriptor instead.
func (*ValidateOTPResponse) Descriptor() ([]byte, []int) {
	return file_proto_otp_proto_rawDescGZIP(), []int{9}
}

func (x *ValidateOTPResponse) GetValid() bool {
	if x != nil {
		return x.Valid
	}
	return false
}

var File_proto_otp_proto protoreflect.FileDescriptor

var file_proto_otp_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x74, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x12, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x66, 0x0a, 0x10,
	0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x4f, 0x54, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x69, 0x73, 0x73,
	0x75, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x69, 0x73, 0x73, 0x75, 0x65,
	0x72, 0x12, 0x21, 0x0a, 0x0c, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x4e, 0x61, 0x6d, 0x65, 0x22, 0x68, 0x0a, 0x11, 0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x4f, 0x54,
	0x50, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x63,
	0x72, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x65, 0x63, 0x72, 0x65,
	0x74, 0x12, 0x25, 0x0a, 0x0e, 0x72, 0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x5f, 0x63, 0x6f,
	0x64, 0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x72, 0x65, 0x63, 0x6f, 0x76,
	0x65, 0x72, 0x79, 0x43, 0x6f, 0x64, 0x65, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x69, 0x6d, 0x61, 0x67,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x22, 0x65,
	0x0a, 0x16, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x4f, 0x54, 0x50, 0x43, 0x6f, 0x64,
	0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x32, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x68,
	0x6f, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x4f, 0x54, 0x50, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x2e, 0x4f, 0x54, 0x50, 0x4d, 0x65,
	0x74, 0x68, 0x6f, 0x64, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x17, 0x0a, 0x07,
	0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75,
	0x73, 0x65, 0x72, 0x49, 0x64, 0x22, 0x19, 0x0a, 0x17, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74,
	0x65, 0x4f, 0x54, 0x50, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x51, 0x0a, 0x1a, 0x46, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x4f, 0x54, 0x50, 0x45, 0x6e, 0x72,
	0x6f, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17,
	0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x61, 0x73, 0x73, 0x63,
	0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x61, 0x73, 0x73, 0x63,
	0x6f, 0x64, 0x65, 0x22, 0x1d, 0x0a, 0x1b, 0x46, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x4f, 0x54, 0x50,
	0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x4b, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x4f, 0x54, 0x50, 0x49, 0x6e, 0x66, 0x6f,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64,
	0x12, 0x1d, 0x0a, 0x0a, 0x6f, 0x74, 0x70, 0x5f, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6f, 0x74, 0x70, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x22,
	0x32, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x4f, 0x54, 0x50, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x72, 0x65, 0x63, 0x69, 0x70, 0x69, 0x65,
	0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x63, 0x69, 0x70, 0x69,
	0x65, 0x6e, 0x74, 0x22, 0x49, 0x0a, 0x12, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x4f,
	0x54, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x61, 0x73,
	0x73, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x61, 0x73,
	0x73, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x22, 0x2b,
	0x0a, 0x13, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x4f, 0x54, 0x50, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x32, 0x8f, 0x03, 0x0a, 0x0a,
	0x4f, 0x54, 0x50, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x43, 0x0a, 0x0a, 0x47, 0x65,
	0x74, 0x4f, 0x54, 0x50, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x18, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x47, 0x65, 0x74, 0x4f, 0x54, 0x50, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x19, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x47, 0x65, 0x74, 0x4f, 0x54,
	0x50, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12,
	0x40, 0x0a, 0x09, 0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x4f, 0x54, 0x50, 0x12, 0x17, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x4f, 0x54, 0x50, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x45, 0x6e,
	0x72, 0x6f, 0x6c, 0x6c, 0x4f, 0x54, 0x50, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x12, 0x52, 0x0a, 0x0f, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x4f, 0x54, 0x50,
	0x43, 0x6f, 0x64, 0x65, 0x12, 0x1d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x47, 0x65, 0x6e,
	0x65, 0x72, 0x61, 0x74, 0x65, 0x4f, 0x54, 0x50, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x47, 0x65, 0x6e, 0x65,
	0x72, 0x61, 0x74, 0x65, 0x4f, 0x54, 0x50, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x5e, 0x0a, 0x13, 0x46, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x4f,
	0x54, 0x50, 0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x21, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x46, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x4f, 0x54, 0x50, 0x45, 0x6e,
	0x72, 0x6f, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x22, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x46, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x4f, 0x54,
	0x50, 0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x46, 0x0a, 0x0b, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x65, 0x4f, 0x54, 0x50, 0x12, 0x19, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x56, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x65, 0x4f, 0x54, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x4f, 0x54, 0x50, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x0c, 0x5a,
	0x0a, 0x63, 0x63, 0x68, 0x6f, 0x69, 0x63, 0x65, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_proto_otp_proto_rawDescOnce sync.Once
	file_proto_otp_proto_rawDescData = file_proto_otp_proto_rawDesc
)

func file_proto_otp_proto_rawDescGZIP() []byte {
	file_proto_otp_proto_rawDescOnce.Do(func() {
		file_proto_otp_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_otp_proto_rawDescData)
	})
	return file_proto_otp_proto_rawDescData
}

var file_proto_otp_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_proto_otp_proto_goTypes = []interface{}{
	(*EnrollOTPRequest)(nil),            // 0: proto.EnrollOTPRequest
	(*EnrollOTPResponse)(nil),           // 1: proto.EnrollOTPResponse
	(*GenerateOTPCodeRequest)(nil),      // 2: proto.GenerateOTPCodeRequest
	(*GenerateOTPCodeResponse)(nil),     // 3: proto.GenerateOTPCodeResponse
	(*FinishOTPEnrollmentRequest)(nil),  // 4: proto.FinishOTPEnrollmentRequest
	(*FinishOTPEnrollmentResponse)(nil), // 5: proto.FinishOTPEnrollmentResponse
	(*GetOTPInfoRequest)(nil),           // 6: proto.GetOTPInfoRequest
	(*GetOTPInfoResponse)(nil),          // 7: proto.GetOTPInfoResponse
	(*ValidateOTPRequest)(nil),          // 8: proto.ValidateOTPRequest
	(*ValidateOTPResponse)(nil),         // 9: proto.ValidateOTPResponse
	(OTPMethod_OTPMethod)(0),            // 10: proto.OTPMethod.OTPMethod
}
var file_proto_otp_proto_depIdxs = []int32{
	10, // 0: proto.GenerateOTPCodeRequest.method:type_name -> proto.OTPMethod.OTPMethod
	6,  // 1: proto.OTPService.GetOTPInfo:input_type -> proto.GetOTPInfoRequest
	0,  // 2: proto.OTPService.EnrollOTP:input_type -> proto.EnrollOTPRequest
	2,  // 3: proto.OTPService.GenerateOTPCode:input_type -> proto.GenerateOTPCodeRequest
	4,  // 4: proto.OTPService.FinishOTPEnrollment:input_type -> proto.FinishOTPEnrollmentRequest
	8,  // 5: proto.OTPService.ValidateOTP:input_type -> proto.ValidateOTPRequest
	7,  // 6: proto.OTPService.GetOTPInfo:output_type -> proto.GetOTPInfoResponse
	1,  // 7: proto.OTPService.EnrollOTP:output_type -> proto.EnrollOTPResponse
	3,  // 8: proto.OTPService.GenerateOTPCode:output_type -> proto.GenerateOTPCodeResponse
	5,  // 9: proto.OTPService.FinishOTPEnrollment:output_type -> proto.FinishOTPEnrollmentResponse
	9,  // 10: proto.OTPService.ValidateOTP:output_type -> proto.ValidateOTPResponse
	6,  // [6:11] is the sub-list for method output_type
	1,  // [1:6] is the sub-list for method input_type
	1,  // [1:1] is the sub-list for extension type_name
	1,  // [1:1] is the sub-list for extension extendee
	0,  // [0:1] is the sub-list for field type_name
}

func init() { file_proto_otp_proto_init() }
func file_proto_otp_proto_init() {
	if File_proto_otp_proto != nil {
		return
	}
	file_proto_common_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_proto_otp_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EnrollOTPRequest); i {
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
		file_proto_otp_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EnrollOTPResponse); i {
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
		file_proto_otp_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GenerateOTPCodeRequest); i {
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
		file_proto_otp_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GenerateOTPCodeResponse); i {
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
		file_proto_otp_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FinishOTPEnrollmentRequest); i {
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
		file_proto_otp_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FinishOTPEnrollmentResponse); i {
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
		file_proto_otp_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetOTPInfoRequest); i {
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
		file_proto_otp_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetOTPInfoResponse); i {
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
		file_proto_otp_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidateOTPRequest); i {
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
		file_proto_otp_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidateOTPResponse); i {
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
			RawDescriptor: file_proto_otp_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_otp_proto_goTypes,
		DependencyIndexes: file_proto_otp_proto_depIdxs,
		MessageInfos:      file_proto_otp_proto_msgTypes,
	}.Build()
	File_proto_otp_proto = out.File
	file_proto_otp_proto_rawDesc = nil
	file_proto_otp_proto_goTypes = nil
	file_proto_otp_proto_depIdxs = nil
}
