// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        (unknown)
// source: proto/validate/v1/validate.proto

package v1

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CheckRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Validations []*Validation `protobuf:"bytes,1,rep,name=validations,proto3" json:"validations,omitempty"`
}

func (x *CheckRequest) Reset() {
	*x = CheckRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_validate_v1_validate_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckRequest) ProtoMessage() {}

func (x *CheckRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_validate_v1_validate_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckRequest.ProtoReflect.Descriptor instead.
func (*CheckRequest) Descriptor() ([]byte, []int) {
	return file_proto_validate_v1_validate_proto_rawDescGZIP(), []int{0}
}

func (x *CheckRequest) GetValidations() []*Validation {
	if x != nil {
		return x.Validations
	}
	return nil
}

type Validation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        string                `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Variables map[string]*anypb.Any `protobuf:"bytes,2,rep,name=variables,proto3" json:"variables,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Validation) Reset() {
	*x = Validation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_validate_v1_validate_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Validation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Validation) ProtoMessage() {}

func (x *Validation) ProtoReflect() protoreflect.Message {
	mi := &file_proto_validate_v1_validate_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Validation.ProtoReflect.Descriptor instead.
func (*Validation) Descriptor() ([]byte, []int) {
	return file_proto_validate_v1_validate_proto_rawDescGZIP(), []int{1}
}

func (x *Validation) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Validation) GetVariables() map[string]*anypb.Any {
	if x != nil {
		return x.Variables
	}
	return nil
}

type CheckResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Results []*ValidationResult `protobuf:"bytes,1,rep,name=results,proto3" json:"results,omitempty"`
}

func (x *CheckResponse) Reset() {
	*x = CheckResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_validate_v1_validate_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckResponse) ProtoMessage() {}

func (x *CheckResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_validate_v1_validate_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckResponse.ProtoReflect.Descriptor instead.
func (*CheckResponse) Descriptor() ([]byte, []int) {
	return file_proto_validate_v1_validate_proto_rawDescGZIP(), []int{2}
}

func (x *CheckResponse) GetResults() []*ValidationResult {
	if x != nil {
		return x.Results
	}
	return nil
}

type ValidationResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	IsValid bool   `protobuf:"varint,2,opt,name=is_valid,json=isValid,proto3" json:"is_valid,omitempty"`
	Message string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *ValidationResult) Reset() {
	*x = ValidationResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_validate_v1_validate_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidationResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidationResult) ProtoMessage() {}

func (x *ValidationResult) ProtoReflect() protoreflect.Message {
	mi := &file_proto_validate_v1_validate_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidationResult.ProtoReflect.Descriptor instead.
func (*ValidationResult) Descriptor() ([]byte, []int) {
	return file_proto_validate_v1_validate_proto_rawDescGZIP(), []int{3}
}

func (x *ValidationResult) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ValidationResult) GetIsValid() bool {
	if x != nil {
		return x.IsValid
	}
	return false
}

func (x *ValidationResult) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_proto_validate_v1_validate_proto protoreflect.FileDescriptor

var file_proto_validate_v1_validate_proto_rawDesc = []byte{
	0x0a, 0x20, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x2f, 0x76, 0x31, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0b, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x1a,
	0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f,
	0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x32, 0x2f,
	0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xee, 0x01, 0x0a, 0x0c, 0x43, 0x68,
	0x65, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0xdd, 0x01, 0x0a, 0x0b, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x17, 0x2e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x56,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0xa1, 0x01, 0x92, 0x41, 0x9a, 0x01,
	0x4a, 0x97, 0x01, 0x5b, 0x7b, 0x22, 0x69, 0x64, 0x22, 0x3a, 0x20, 0x22, 0x69, 0x74, 0x65, 0x6d,
	0x22, 0x2c, 0x20, 0x22, 0x76, 0x61, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x22, 0x3a, 0x20,
	0x7b, 0x22, 0x70, 0x72, 0x69, 0x63, 0x65, 0x22, 0x3a, 0x20, 0x2d, 0x31, 0x30, 0x30, 0x2c, 0x20,
	0x22, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x22, 0x3a, 0x20, 0x22, 0x69, 0x56, 0x42, 0x4f, 0x52, 0x77,
	0x30, 0x4b, 0x47, 0x67, 0x6f, 0x41, 0x41, 0x41, 0x41, 0x4e, 0x53, 0x55, 0x68, 0x45, 0x55, 0x67,
	0x41, 0x41, 0x41, 0x41, 0x45, 0x41, 0x41, 0x41, 0x41, 0x42, 0x43, 0x41, 0x49, 0x41, 0x41, 0x41,
	0x43, 0x51, 0x64, 0x31, 0x50, 0x65, 0x41, 0x41, 0x41, 0x41, 0x44, 0x45, 0x6c, 0x45, 0x51, 0x56,
	0x52, 0x34, 0x6e, 0x47, 0x4f, 0x34, 0x75, 0x6e, 0x59, 0x32, 0x41, 0x41, 0x52, 0x34, 0x41, 0x68,
	0x35, 0x31, 0x6a, 0x35, 0x58, 0x77, 0x41, 0x41, 0x41, 0x41, 0x41, 0x45, 0x6c, 0x46, 0x54, 0x6b,
	0x53, 0x75, 0x51, 0x6d, 0x43, 0x43, 0x22, 0x7d, 0x7d, 0x5d, 0xe0, 0x41, 0x02, 0x52, 0x0b, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0xc0, 0x01, 0x0a, 0x0a, 0x56,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x13, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x02, 0x69, 0x64, 0x12, 0x49,
	0x0a, 0x09, 0x76, 0x61, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x26, 0x2e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e,
	0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x56, 0x61, 0x72, 0x69, 0x61,
	0x62, 0x6c, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x09,
	0x76, 0x61, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x1a, 0x52, 0x0a, 0x0e, 0x56, 0x61, 0x72,
	0x69, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2a, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41,
	0x6e, 0x79, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xa2, 0x01,
	0x0a, 0x0d, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x90, 0x01, 0x0a, 0x07, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x1d, 0x2e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e,
	0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x42, 0x57, 0x92, 0x41, 0x51, 0x4a, 0x4f, 0x5b, 0x7b, 0x22, 0x69, 0x64, 0x22, 0x3a, 0x20, 0x22,
	0x69, 0x74, 0x65, 0x6d, 0x22, 0x2c, 0x20, 0x22, 0x69, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x69, 0x64,
	0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x20, 0x22, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x22, 0x3a, 0x20, 0x22, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x20, 0x76, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x3a, 0x20, 0x70, 0x72, 0x69, 0x63, 0x65, 0x20,
	0x3e, 0x20, 0x30, 0x22, 0x7d, 0x5d, 0xe0, 0x41, 0x02, 0x52, 0x07, 0x72, 0x65, 0x73, 0x75, 0x6c,
	0x74, 0x73, 0x22, 0x66, 0x0a, 0x10, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x13, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1e, 0x0a, 0x08, 0x69,
	0x73, 0x5f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x42, 0x03, 0xe0,
	0x41, 0x02, 0x52, 0x07, 0x69, 0x73, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x12, 0x1d, 0x0a, 0x07, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41,
	0x02, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x32, 0x8f, 0x01, 0x0a, 0x0f, 0x56,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x7c,
	0x0a, 0x05, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x12, 0x19, 0x2e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31,
	0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x3c,
	0x92, 0x41, 0x25, 0x0a, 0x0a, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x10, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x20, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x2a, 0x05, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0e, 0x3a, 0x01,
	0x2a, 0x22, 0x09, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x42, 0x13, 0x5a, 0x11,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76,
	0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_validate_v1_validate_proto_rawDescOnce sync.Once
	file_proto_validate_v1_validate_proto_rawDescData = file_proto_validate_v1_validate_proto_rawDesc
)

func file_proto_validate_v1_validate_proto_rawDescGZIP() []byte {
	file_proto_validate_v1_validate_proto_rawDescOnce.Do(func() {
		file_proto_validate_v1_validate_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_validate_v1_validate_proto_rawDescData)
	})
	return file_proto_validate_v1_validate_proto_rawDescData
}

var file_proto_validate_v1_validate_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_proto_validate_v1_validate_proto_goTypes = []interface{}{
	(*CheckRequest)(nil),     // 0: validate.v1.CheckRequest
	(*Validation)(nil),       // 1: validate.v1.Validation
	(*CheckResponse)(nil),    // 2: validate.v1.CheckResponse
	(*ValidationResult)(nil), // 3: validate.v1.ValidationResult
	nil,                      // 4: validate.v1.Validation.VariablesEntry
	(*anypb.Any)(nil),        // 5: google.protobuf.Any
}
var file_proto_validate_v1_validate_proto_depIdxs = []int32{
	1, // 0: validate.v1.CheckRequest.validations:type_name -> validate.v1.Validation
	4, // 1: validate.v1.Validation.variables:type_name -> validate.v1.Validation.VariablesEntry
	3, // 2: validate.v1.CheckResponse.results:type_name -> validate.v1.ValidationResult
	5, // 3: validate.v1.Validation.VariablesEntry.value:type_name -> google.protobuf.Any
	0, // 4: validate.v1.ValidateService.Check:input_type -> validate.v1.CheckRequest
	2, // 5: validate.v1.ValidateService.Check:output_type -> validate.v1.CheckResponse
	5, // [5:6] is the sub-list for method output_type
	4, // [4:5] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_proto_validate_v1_validate_proto_init() }
func file_proto_validate_v1_validate_proto_init() {
	if File_proto_validate_v1_validate_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_validate_v1_validate_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckRequest); i {
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
		file_proto_validate_v1_validate_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Validation); i {
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
		file_proto_validate_v1_validate_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckResponse); i {
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
		file_proto_validate_v1_validate_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidationResult); i {
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
			RawDescriptor: file_proto_validate_v1_validate_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_validate_v1_validate_proto_goTypes,
		DependencyIndexes: file_proto_validate_v1_validate_proto_depIdxs,
		MessageInfos:      file_proto_validate_v1_validate_proto_msgTypes,
	}.Build()
	File_proto_validate_v1_validate_proto = out.File
	file_proto_validate_v1_validate_proto_rawDesc = nil
	file_proto_validate_v1_validate_proto_goTypes = nil
	file_proto_validate_v1_validate_proto_depIdxs = nil
}
