// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.3
// source: kvstore.proto

package proto

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

// KVStoreGet is a structure used to create Get request callbacks to the Tarmac
// KVStore interface.
//
// This structure is a general request type used for all KVStore types provided
// by Tarmac.
type KVStoreGet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Key is the index key to use when accessing the key:value store.
	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *KVStoreGet) Reset() {
	*x = KVStoreGet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kvstore_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVStoreGet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVStoreGet) ProtoMessage() {}

func (x *KVStoreGet) ProtoReflect() protoreflect.Message {
	mi := &file_kvstore_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVStoreGet.ProtoReflect.Descriptor instead.
func (*KVStoreGet) Descriptor() ([]byte, []int) {
	return file_kvstore_proto_rawDescGZIP(), []int{0}
}

func (x *KVStoreGet) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

// KVStoreGetResponse is a structure supplied as response messages to KVStore
// Get requests.
//
// This response is a general response type used for all KVStore types provided
// by Tarmac.
type KVStoreGetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Status is the human readable error message or success message for function
	// execution.
	Status *Status `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	// Data is the response data provided by the key:value store.
	// This data is a byte slice to provide a simple field for arbitrary data.
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *KVStoreGetResponse) Reset() {
	*x = KVStoreGetResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kvstore_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVStoreGetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVStoreGetResponse) ProtoMessage() {}

func (x *KVStoreGetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kvstore_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVStoreGetResponse.ProtoReflect.Descriptor instead.
func (*KVStoreGetResponse) Descriptor() ([]byte, []int) {
	return file_kvstore_proto_rawDescGZIP(), []int{1}
}

func (x *KVStoreGetResponse) GetStatus() *Status {
	if x != nil {
		return x.Status
	}
	return nil
}

func (x *KVStoreGetResponse) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

// KVStoreSet is a structure used to create a Set request callback to the
// Tarmac KVStore interface.
//
// This structure is a general request type used for all KVStore types provided
// by Tarmac.
type KVStoreSet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Key is the index key used to store the data.
	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	// Data is the user-supplied key:value data.
	// Tarmac expects this field to be a byte slice.
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *KVStoreSet) Reset() {
	*x = KVStoreSet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kvstore_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVStoreSet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVStoreSet) ProtoMessage() {}

func (x *KVStoreSet) ProtoReflect() protoreflect.Message {
	mi := &file_kvstore_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVStoreSet.ProtoReflect.Descriptor instead.
func (*KVStoreSet) Descriptor() ([]byte, []int) {
	return file_kvstore_proto_rawDescGZIP(), []int{2}
}

func (x *KVStoreSet) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *KVStoreSet) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

// KVStoreSetResponse is a structure supplied as a response message to the
// KVStore Set callback function.
//
// This response is a general response type used for all KVStore types provided
// by Tarmac.
type KVStoreSetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Status is the human readable error message or success message for function
	// execution.
	Status *Status `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *KVStoreSetResponse) Reset() {
	*x = KVStoreSetResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kvstore_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVStoreSetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVStoreSetResponse) ProtoMessage() {}

func (x *KVStoreSetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kvstore_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVStoreSetResponse.ProtoReflect.Descriptor instead.
func (*KVStoreSetResponse) Descriptor() ([]byte, []int) {
	return file_kvstore_proto_rawDescGZIP(), []int{3}
}

func (x *KVStoreSetResponse) GetStatus() *Status {
	if x != nil {
		return x.Status
	}
	return nil
}

// KVStoreDelete is a structure used to create Delete callback requests to the
// Tarmac KVStore interface.
//
// This structure is a general request type used for all KVStore types provided
// by Tarmac.
type KVStoreDelete struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Key is the index key used to store the data.
	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *KVStoreDelete) Reset() {
	*x = KVStoreDelete{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kvstore_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVStoreDelete) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVStoreDelete) ProtoMessage() {}

func (x *KVStoreDelete) ProtoReflect() protoreflect.Message {
	mi := &file_kvstore_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVStoreDelete.ProtoReflect.Descriptor instead.
func (*KVStoreDelete) Descriptor() ([]byte, []int) {
	return file_kvstore_proto_rawDescGZIP(), []int{4}
}

func (x *KVStoreDelete) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

// KVStoreDeleteResponse is a structure supplied as a response message to the
// KVStore Delete callback function.
//
// This response is a general response type used for all KVStore types provided
// by Tarmac.
type KVStoreDeleteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Status is the human readable error message or success message for function
	// execution.
	Status *Status `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *KVStoreDeleteResponse) Reset() {
	*x = KVStoreDeleteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kvstore_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVStoreDeleteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVStoreDeleteResponse) ProtoMessage() {}

func (x *KVStoreDeleteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kvstore_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVStoreDeleteResponse.ProtoReflect.Descriptor instead.
func (*KVStoreDeleteResponse) Descriptor() ([]byte, []int) {
	return file_kvstore_proto_rawDescGZIP(), []int{5}
}

func (x *KVStoreDeleteResponse) GetStatus() *Status {
	if x != nil {
		return x.Status
	}
	return nil
}

// KVStoreKeys is a structure used to create a Keys callback request to the
// Tarmac KVStore interface.
//
// This structure is a general request type used for all KVStore types provided
// by Tarmac.
type KVStoreKeys struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ReturnProto is a boolean value that determines if the response should be
	// returned as a JSON string or as a KVStoreKeysResponse message.
	//
	// This must be set to true to return a KVStoreKeysResponse message.
	ReturnProto bool `protobuf:"varint,1,opt,name=return_proto,json=returnProto,proto3" json:"return_proto,omitempty"`
}

func (x *KVStoreKeys) Reset() {
	*x = KVStoreKeys{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kvstore_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVStoreKeys) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVStoreKeys) ProtoMessage() {}

func (x *KVStoreKeys) ProtoReflect() protoreflect.Message {
	mi := &file_kvstore_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVStoreKeys.ProtoReflect.Descriptor instead.
func (*KVStoreKeys) Descriptor() ([]byte, []int) {
	return file_kvstore_proto_rawDescGZIP(), []int{6}
}

func (x *KVStoreKeys) GetReturnProto() bool {
	if x != nil {
		return x.ReturnProto
	}
	return false
}

// KVStoreKeysResponse is a structure supplied as a response message to the
// KVStore Keys callback function.
//
// This response is a general response type used for all KVStore types provided
// by Tarmac.
type KVStoreKeysResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Status is the human readable error message or success message for function
	// execution.
	Status *Status `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	// Keys is a list of keys available within the KV Store.
	Keys []string `protobuf:"bytes,2,rep,name=keys,proto3" json:"keys,omitempty"`
}

func (x *KVStoreKeysResponse) Reset() {
	*x = KVStoreKeysResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kvstore_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVStoreKeysResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVStoreKeysResponse) ProtoMessage() {}

func (x *KVStoreKeysResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kvstore_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVStoreKeysResponse.ProtoReflect.Descriptor instead.
func (*KVStoreKeysResponse) Descriptor() ([]byte, []int) {
	return file_kvstore_proto_rawDescGZIP(), []int{7}
}

func (x *KVStoreKeysResponse) GetStatus() *Status {
	if x != nil {
		return x.Status
	}
	return nil
}

func (x *KVStoreKeysResponse) GetKeys() []string {
	if x != nil {
		return x.Keys
	}
	return nil
}

var File_kvstore_proto protoreflect.FileDescriptor

var file_kvstore_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x6b, 0x76, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x0e, 0x74, 0x61, 0x72, 0x6d, 0x61, 0x63, 0x2e, 0x6b, 0x76, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x1a,
	0x0c, 0x74, 0x61, 0x72, 0x6d, 0x61, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x1e, 0x0a,
	0x0a, 0x4b, 0x56, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x47, 0x65, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x22, 0x50, 0x0a,
	0x12, 0x4b, 0x56, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x26, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x74, 0x61, 0x72, 0x6d, 0x61, 0x63, 0x2e, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22,
	0x32, 0x0a, 0x0a, 0x4b, 0x56, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x53, 0x65, 0x74, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x22, 0x3c, 0x0a, 0x12, 0x4b, 0x56, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x53, 0x65,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x26, 0x0a, 0x06, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x74, 0x61, 0x72, 0x6d,
	0x61, 0x63, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x22, 0x21, 0x0a, 0x0d, 0x4b, 0x56, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x44, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x22, 0x3f, 0x0a, 0x15, 0x4b, 0x56, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x44,
	0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x26, 0x0a,
	0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e,
	0x74, 0x61, 0x72, 0x6d, 0x61, 0x63, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x30, 0x0a, 0x0b, 0x4b, 0x56, 0x53, 0x74, 0x6f, 0x72, 0x65,
	0x4b, 0x65, 0x79, 0x73, 0x12, 0x21, 0x0a, 0x0c, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x5f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x72, 0x65, 0x74, 0x75,
	0x72, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x51, 0x0a, 0x13, 0x4b, 0x56, 0x53, 0x74, 0x6f,
	0x72, 0x65, 0x4b, 0x65, 0x79, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x26,
	0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e,
	0x2e, 0x74, 0x61, 0x72, 0x6d, 0x61, 0x63, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x42, 0x28, 0x5a, 0x26, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x61, 0x72, 0x6d, 0x61, 0x63, 0x2d,
	0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x74, 0x61, 0x72, 0x6d, 0x61, 0x63, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kvstore_proto_rawDescOnce sync.Once
	file_kvstore_proto_rawDescData = file_kvstore_proto_rawDesc
)

func file_kvstore_proto_rawDescGZIP() []byte {
	file_kvstore_proto_rawDescOnce.Do(func() {
		file_kvstore_proto_rawDescData = protoimpl.X.CompressGZIP(file_kvstore_proto_rawDescData)
	})
	return file_kvstore_proto_rawDescData
}

var file_kvstore_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_kvstore_proto_goTypes = []any{
	(*KVStoreGet)(nil),            // 0: tarmac.kvstore.KVStoreGet
	(*KVStoreGetResponse)(nil),    // 1: tarmac.kvstore.KVStoreGetResponse
	(*KVStoreSet)(nil),            // 2: tarmac.kvstore.KVStoreSet
	(*KVStoreSetResponse)(nil),    // 3: tarmac.kvstore.KVStoreSetResponse
	(*KVStoreDelete)(nil),         // 4: tarmac.kvstore.KVStoreDelete
	(*KVStoreDeleteResponse)(nil), // 5: tarmac.kvstore.KVStoreDeleteResponse
	(*KVStoreKeys)(nil),           // 6: tarmac.kvstore.KVStoreKeys
	(*KVStoreKeysResponse)(nil),   // 7: tarmac.kvstore.KVStoreKeysResponse
	(*Status)(nil),                // 8: tarmac.Status
}
var file_kvstore_proto_depIdxs = []int32{
	8, // 0: tarmac.kvstore.KVStoreGetResponse.status:type_name -> tarmac.Status
	8, // 1: tarmac.kvstore.KVStoreSetResponse.status:type_name -> tarmac.Status
	8, // 2: tarmac.kvstore.KVStoreDeleteResponse.status:type_name -> tarmac.Status
	8, // 3: tarmac.kvstore.KVStoreKeysResponse.status:type_name -> tarmac.Status
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_kvstore_proto_init() }
func file_kvstore_proto_init() {
	if File_kvstore_proto != nil {
		return
	}
	file_tarmac_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_kvstore_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*KVStoreGet); i {
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
		file_kvstore_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*KVStoreGetResponse); i {
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
		file_kvstore_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*KVStoreSet); i {
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
		file_kvstore_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*KVStoreSetResponse); i {
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
		file_kvstore_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*KVStoreDelete); i {
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
		file_kvstore_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*KVStoreDeleteResponse); i {
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
		file_kvstore_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*KVStoreKeys); i {
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
		file_kvstore_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*KVStoreKeysResponse); i {
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
			RawDescriptor: file_kvstore_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kvstore_proto_goTypes,
		DependencyIndexes: file_kvstore_proto_depIdxs,
		MessageInfos:      file_kvstore_proto_msgTypes,
	}.Build()
	File_kvstore_proto = out.File
	file_kvstore_proto_rawDesc = nil
	file_kvstore_proto_goTypes = nil
	file_kvstore_proto_depIdxs = nil
}
