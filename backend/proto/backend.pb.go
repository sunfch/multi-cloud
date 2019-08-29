// Code generated by protoc-gen-go. DO NOT EDIT.
// source: backend.proto

package backend

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type CreateBackendRequest struct {
	Backend              *BackendDetail `protobuf:"bytes,1,opt,name=backend,proto3" json:"backend,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *CreateBackendRequest) Reset()         { *m = CreateBackendRequest{} }
func (m *CreateBackendRequest) String() string { return proto.CompactTextString(m) }
func (*CreateBackendRequest) ProtoMessage()    {}
func (*CreateBackendRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_backend_89cca7ce8354284a, []int{0}
}
func (m *CreateBackendRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateBackendRequest.Unmarshal(m, b)
}
func (m *CreateBackendRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateBackendRequest.Marshal(b, m, deterministic)
}
func (dst *CreateBackendRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateBackendRequest.Merge(dst, src)
}
func (m *CreateBackendRequest) XXX_Size() int {
	return xxx_messageInfo_CreateBackendRequest.Size(m)
}
func (m *CreateBackendRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateBackendRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CreateBackendRequest proto.InternalMessageInfo

func (m *CreateBackendRequest) GetBackend() *BackendDetail {
	if m != nil {
		return m.Backend
	}
	return nil
}

type CreateBackendResponse struct {
	Backend              *BackendDetail `protobuf:"bytes,1,opt,name=backend,proto3" json:"backend,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *CreateBackendResponse) Reset()         { *m = CreateBackendResponse{} }
func (m *CreateBackendResponse) String() string { return proto.CompactTextString(m) }
func (*CreateBackendResponse) ProtoMessage()    {}
func (*CreateBackendResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_backend_89cca7ce8354284a, []int{1}
}
func (m *CreateBackendResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateBackendResponse.Unmarshal(m, b)
}
func (m *CreateBackendResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateBackendResponse.Marshal(b, m, deterministic)
}
func (dst *CreateBackendResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateBackendResponse.Merge(dst, src)
}
func (m *CreateBackendResponse) XXX_Size() int {
	return xxx_messageInfo_CreateBackendResponse.Size(m)
}
func (m *CreateBackendResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateBackendResponse.DiscardUnknown(m)
}

var xxx_messageInfo_CreateBackendResponse proto.InternalMessageInfo

func (m *CreateBackendResponse) GetBackend() *BackendDetail {
	if m != nil {
		return m.Backend
	}
	return nil
}

type GetBackendRequest struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetBackendRequest) Reset()         { *m = GetBackendRequest{} }
func (m *GetBackendRequest) String() string { return proto.CompactTextString(m) }
func (*GetBackendRequest) ProtoMessage()    {}
func (*GetBackendRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_backend_89cca7ce8354284a, []int{2}
}
func (m *GetBackendRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetBackendRequest.Unmarshal(m, b)
}
func (m *GetBackendRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetBackendRequest.Marshal(b, m, deterministic)
}
func (dst *GetBackendRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetBackendRequest.Merge(dst, src)
}
func (m *GetBackendRequest) XXX_Size() int {
	return xxx_messageInfo_GetBackendRequest.Size(m)
}
func (m *GetBackendRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetBackendRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetBackendRequest proto.InternalMessageInfo

func (m *GetBackendRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type GetBackendResponse struct {
	Backend              *BackendDetail `protobuf:"bytes,1,opt,name=backend,proto3" json:"backend,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *GetBackendResponse) Reset()         { *m = GetBackendResponse{} }
func (m *GetBackendResponse) String() string { return proto.CompactTextString(m) }
func (*GetBackendResponse) ProtoMessage()    {}
func (*GetBackendResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_backend_89cca7ce8354284a, []int{3}
}
func (m *GetBackendResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetBackendResponse.Unmarshal(m, b)
}
func (m *GetBackendResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetBackendResponse.Marshal(b, m, deterministic)
}
func (dst *GetBackendResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetBackendResponse.Merge(dst, src)
}
func (m *GetBackendResponse) XXX_Size() int {
	return xxx_messageInfo_GetBackendResponse.Size(m)
}
func (m *GetBackendResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetBackendResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetBackendResponse proto.InternalMessageInfo

func (m *GetBackendResponse) GetBackend() *BackendDetail {
	if m != nil {
		return m.Backend
	}
	return nil
}

type ListBackendRequest struct {
	Limit                int32             `protobuf:"varint,1,opt,name=limit,proto3" json:"limit,omitempty"`
	Offset               int32             `protobuf:"varint,2,opt,name=offset,proto3" json:"offset,omitempty"`
	SortKeys             []string          `protobuf:"bytes,3,rep,name=sortKeys,proto3" json:"sortKeys,omitempty"`
	SortDirs             []string          `protobuf:"bytes,4,rep,name=sortDirs,proto3" json:"sortDirs,omitempty"`
	Filter               map[string]string `protobuf:"bytes,5,rep,name=Filter,proto3" json:"Filter,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *ListBackendRequest) Reset()         { *m = ListBackendRequest{} }
func (m *ListBackendRequest) String() string { return proto.CompactTextString(m) }
func (*ListBackendRequest) ProtoMessage()    {}
func (*ListBackendRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_backend_89cca7ce8354284a, []int{4}
}
func (m *ListBackendRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListBackendRequest.Unmarshal(m, b)
}
func (m *ListBackendRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListBackendRequest.Marshal(b, m, deterministic)
}
func (dst *ListBackendRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListBackendRequest.Merge(dst, src)
}
func (m *ListBackendRequest) XXX_Size() int {
	return xxx_messageInfo_ListBackendRequest.Size(m)
}
func (m *ListBackendRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ListBackendRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ListBackendRequest proto.InternalMessageInfo

func (m *ListBackendRequest) GetLimit() int32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

func (m *ListBackendRequest) GetOffset() int32 {
	if m != nil {
		return m.Offset
	}
	return 0
}

func (m *ListBackendRequest) GetSortKeys() []string {
	if m != nil {
		return m.SortKeys
	}
	return nil
}

func (m *ListBackendRequest) GetSortDirs() []string {
	if m != nil {
		return m.SortDirs
	}
	return nil
}

func (m *ListBackendRequest) GetFilter() map[string]string {
	if m != nil {
		return m.Filter
	}
	return nil
}

type ListBackendResponse struct {
	Backends             []*BackendDetail `protobuf:"bytes,1,rep,name=backends,proto3" json:"backends,omitempty"`
	Next                 int32            `protobuf:"varint,2,opt,name=next,proto3" json:"next,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *ListBackendResponse) Reset()         { *m = ListBackendResponse{} }
func (m *ListBackendResponse) String() string { return proto.CompactTextString(m) }
func (*ListBackendResponse) ProtoMessage()    {}
func (*ListBackendResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_backend_89cca7ce8354284a, []int{5}
}
func (m *ListBackendResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListBackendResponse.Unmarshal(m, b)
}
func (m *ListBackendResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListBackendResponse.Marshal(b, m, deterministic)
}
func (dst *ListBackendResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListBackendResponse.Merge(dst, src)
}
func (m *ListBackendResponse) XXX_Size() int {
	return xxx_messageInfo_ListBackendResponse.Size(m)
}
func (m *ListBackendResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListBackendResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListBackendResponse proto.InternalMessageInfo

func (m *ListBackendResponse) GetBackends() []*BackendDetail {
	if m != nil {
		return m.Backends
	}
	return nil
}

func (m *ListBackendResponse) GetNext() int32 {
	if m != nil {
		return m.Next
	}
	return 0
}

type UpdateBackendRequest struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Access               string   `protobuf:"bytes,2,opt,name=access,proto3" json:"access,omitempty"`
	Security             string   `protobuf:"bytes,3,opt,name=security,proto3" json:"security,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UpdateBackendRequest) Reset()         { *m = UpdateBackendRequest{} }
func (m *UpdateBackendRequest) String() string { return proto.CompactTextString(m) }
func (*UpdateBackendRequest) ProtoMessage()    {}
func (*UpdateBackendRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_backend_89cca7ce8354284a, []int{6}
}
func (m *UpdateBackendRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateBackendRequest.Unmarshal(m, b)
}
func (m *UpdateBackendRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateBackendRequest.Marshal(b, m, deterministic)
}
func (dst *UpdateBackendRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateBackendRequest.Merge(dst, src)
}
func (m *UpdateBackendRequest) XXX_Size() int {
	return xxx_messageInfo_UpdateBackendRequest.Size(m)
}
func (m *UpdateBackendRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateBackendRequest.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateBackendRequest proto.InternalMessageInfo

func (m *UpdateBackendRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *UpdateBackendRequest) GetAccess() string {
	if m != nil {
		return m.Access
	}
	return ""
}

func (m *UpdateBackendRequest) GetSecurity() string {
	if m != nil {
		return m.Security
	}
	return ""
}

type UpdateBackendResponse struct {
	Backend              *BackendDetail `protobuf:"bytes,1,opt,name=backend,proto3" json:"backend,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *UpdateBackendResponse) Reset()         { *m = UpdateBackendResponse{} }
func (m *UpdateBackendResponse) String() string { return proto.CompactTextString(m) }
func (*UpdateBackendResponse) ProtoMessage()    {}
func (*UpdateBackendResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_backend_89cca7ce8354284a, []int{7}
}
func (m *UpdateBackendResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateBackendResponse.Unmarshal(m, b)
}
func (m *UpdateBackendResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateBackendResponse.Marshal(b, m, deterministic)
}
func (dst *UpdateBackendResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateBackendResponse.Merge(dst, src)
}
func (m *UpdateBackendResponse) XXX_Size() int {
	return xxx_messageInfo_UpdateBackendResponse.Size(m)
}
func (m *UpdateBackendResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateBackendResponse.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateBackendResponse proto.InternalMessageInfo

func (m *UpdateBackendResponse) GetBackend() *BackendDetail {
	if m != nil {
		return m.Backend
	}
	return nil
}

type DeleteBackendRequest struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteBackendRequest) Reset()         { *m = DeleteBackendRequest{} }
func (m *DeleteBackendRequest) String() string { return proto.CompactTextString(m) }
func (*DeleteBackendRequest) ProtoMessage()    {}
func (*DeleteBackendRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_backend_89cca7ce8354284a, []int{8}
}
func (m *DeleteBackendRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteBackendRequest.Unmarshal(m, b)
}
func (m *DeleteBackendRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteBackendRequest.Marshal(b, m, deterministic)
}
func (dst *DeleteBackendRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteBackendRequest.Merge(dst, src)
}
func (m *DeleteBackendRequest) XXX_Size() int {
	return xxx_messageInfo_DeleteBackendRequest.Size(m)
}
func (m *DeleteBackendRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteBackendRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteBackendRequest proto.InternalMessageInfo

func (m *DeleteBackendRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type DeleteBackendResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteBackendResponse) Reset()         { *m = DeleteBackendResponse{} }
func (m *DeleteBackendResponse) String() string { return proto.CompactTextString(m) }
func (*DeleteBackendResponse) ProtoMessage()    {}
func (*DeleteBackendResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_backend_89cca7ce8354284a, []int{9}
}
func (m *DeleteBackendResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteBackendResponse.Unmarshal(m, b)
}
func (m *DeleteBackendResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteBackendResponse.Marshal(b, m, deterministic)
}
func (dst *DeleteBackendResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteBackendResponse.Merge(dst, src)
}
func (m *DeleteBackendResponse) XXX_Size() int {
	return xxx_messageInfo_DeleteBackendResponse.Size(m)
}
func (m *DeleteBackendResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteBackendResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteBackendResponse proto.InternalMessageInfo

type BackendDetail struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	TenantId             string   `protobuf:"bytes,2,opt,name=tenantId,proto3" json:"tenantId,omitempty"`
	UserId               string   `protobuf:"bytes,3,opt,name=userId,proto3" json:"userId,omitempty"`
	Name                 string   `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	Type                 string   `protobuf:"bytes,5,opt,name=type,proto3" json:"type,omitempty"`
	Region               string   `protobuf:"bytes,6,opt,name=region,proto3" json:"region,omitempty"`
	Endpoint             string   `protobuf:"bytes,7,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	BucketName           string   `protobuf:"bytes,8,opt,name=bucketName,proto3" json:"bucketName,omitempty"`
	Access               string   `protobuf:"bytes,9,opt,name=access,proto3" json:"access,omitempty"`
	Security             string   `protobuf:"bytes,10,opt,name=security,proto3" json:"security,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BackendDetail) Reset()         { *m = BackendDetail{} }
func (m *BackendDetail) String() string { return proto.CompactTextString(m) }
func (*BackendDetail) ProtoMessage()    {}
func (*BackendDetail) Descriptor() ([]byte, []int) {
	return fileDescriptor_backend_89cca7ce8354284a, []int{10}
}
func (m *BackendDetail) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BackendDetail.Unmarshal(m, b)
}
func (m *BackendDetail) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BackendDetail.Marshal(b, m, deterministic)
}
func (dst *BackendDetail) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BackendDetail.Merge(dst, src)
}
func (m *BackendDetail) XXX_Size() int {
	return xxx_messageInfo_BackendDetail.Size(m)
}
func (m *BackendDetail) XXX_DiscardUnknown() {
	xxx_messageInfo_BackendDetail.DiscardUnknown(m)
}

var xxx_messageInfo_BackendDetail proto.InternalMessageInfo

func (m *BackendDetail) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *BackendDetail) GetTenantId() string {
	if m != nil {
		return m.TenantId
	}
	return ""
}

func (m *BackendDetail) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *BackendDetail) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *BackendDetail) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *BackendDetail) GetRegion() string {
	if m != nil {
		return m.Region
	}
	return ""
}

func (m *BackendDetail) GetEndpoint() string {
	if m != nil {
		return m.Endpoint
	}
	return ""
}

func (m *BackendDetail) GetBucketName() string {
	if m != nil {
		return m.BucketName
	}
	return ""
}

func (m *BackendDetail) GetAccess() string {
	if m != nil {
		return m.Access
	}
	return ""
}

func (m *BackendDetail) GetSecurity() string {
	if m != nil {
		return m.Security
	}
	return ""
}

type ListTypeRequest struct {
	Limit                int32             `protobuf:"varint,1,opt,name=limit,proto3" json:"limit,omitempty"`
	Offset               int32             `protobuf:"varint,2,opt,name=offset,proto3" json:"offset,omitempty"`
	SortKeys             []string          `protobuf:"bytes,3,rep,name=sortKeys,proto3" json:"sortKeys,omitempty"`
	SortDirs             []string          `protobuf:"bytes,4,rep,name=sortDirs,proto3" json:"sortDirs,omitempty"`
	Filter               map[string]string `protobuf:"bytes,5,rep,name=Filter,proto3" json:"Filter,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *ListTypeRequest) Reset()         { *m = ListTypeRequest{} }
func (m *ListTypeRequest) String() string { return proto.CompactTextString(m) }
func (*ListTypeRequest) ProtoMessage()    {}
func (*ListTypeRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_backend_89cca7ce8354284a, []int{11}
}
func (m *ListTypeRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListTypeRequest.Unmarshal(m, b)
}
func (m *ListTypeRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListTypeRequest.Marshal(b, m, deterministic)
}
func (dst *ListTypeRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListTypeRequest.Merge(dst, src)
}
func (m *ListTypeRequest) XXX_Size() int {
	return xxx_messageInfo_ListTypeRequest.Size(m)
}
func (m *ListTypeRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ListTypeRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ListTypeRequest proto.InternalMessageInfo

func (m *ListTypeRequest) GetLimit() int32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

func (m *ListTypeRequest) GetOffset() int32 {
	if m != nil {
		return m.Offset
	}
	return 0
}

func (m *ListTypeRequest) GetSortKeys() []string {
	if m != nil {
		return m.SortKeys
	}
	return nil
}

func (m *ListTypeRequest) GetSortDirs() []string {
	if m != nil {
		return m.SortDirs
	}
	return nil
}

func (m *ListTypeRequest) GetFilter() map[string]string {
	if m != nil {
		return m.Filter
	}
	return nil
}

type ListTypeResponse struct {
	Types                []*TypeDetail `protobuf:"bytes,1,rep,name=types,proto3" json:"types,omitempty"`
	Next                 int32         `protobuf:"varint,2,opt,name=next,proto3" json:"next,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *ListTypeResponse) Reset()         { *m = ListTypeResponse{} }
func (m *ListTypeResponse) String() string { return proto.CompactTextString(m) }
func (*ListTypeResponse) ProtoMessage()    {}
func (*ListTypeResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_backend_89cca7ce8354284a, []int{12}
}
func (m *ListTypeResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListTypeResponse.Unmarshal(m, b)
}
func (m *ListTypeResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListTypeResponse.Marshal(b, m, deterministic)
}
func (dst *ListTypeResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListTypeResponse.Merge(dst, src)
}
func (m *ListTypeResponse) XXX_Size() int {
	return xxx_messageInfo_ListTypeResponse.Size(m)
}
func (m *ListTypeResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListTypeResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListTypeResponse proto.InternalMessageInfo

func (m *ListTypeResponse) GetTypes() []*TypeDetail {
	if m != nil {
		return m.Types
	}
	return nil
}

func (m *ListTypeResponse) GetNext() int32 {
	if m != nil {
		return m.Next
	}
	return 0
}

type TypeDetail struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Description          string   `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TypeDetail) Reset()         { *m = TypeDetail{} }
func (m *TypeDetail) String() string { return proto.CompactTextString(m) }
func (*TypeDetail) ProtoMessage()    {}
func (*TypeDetail) Descriptor() ([]byte, []int) {
	return fileDescriptor_backend_89cca7ce8354284a, []int{13}
}
func (m *TypeDetail) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TypeDetail.Unmarshal(m, b)
}
func (m *TypeDetail) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TypeDetail.Marshal(b, m, deterministic)
}
func (dst *TypeDetail) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TypeDetail.Merge(dst, src)
}
func (m *TypeDetail) XXX_Size() int {
	return xxx_messageInfo_TypeDetail.Size(m)
}
func (m *TypeDetail) XXX_DiscardUnknown() {
	xxx_messageInfo_TypeDetail.DiscardUnknown(m)
}

var xxx_messageInfo_TypeDetail proto.InternalMessageInfo

func (m *TypeDetail) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *TypeDetail) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func init() {
	proto.RegisterType((*CreateBackendRequest)(nil), "CreateBackendRequest")
	proto.RegisterType((*CreateBackendResponse)(nil), "CreateBackendResponse")
	proto.RegisterType((*GetBackendRequest)(nil), "GetBackendRequest")
	proto.RegisterType((*GetBackendResponse)(nil), "GetBackendResponse")
	proto.RegisterType((*ListBackendRequest)(nil), "ListBackendRequest")
	proto.RegisterMapType((map[string]string)(nil), "ListBackendRequest.FilterEntry")
	proto.RegisterType((*ListBackendResponse)(nil), "ListBackendResponse")
	proto.RegisterType((*UpdateBackendRequest)(nil), "UpdateBackendRequest")
	proto.RegisterType((*UpdateBackendResponse)(nil), "UpdateBackendResponse")
	proto.RegisterType((*DeleteBackendRequest)(nil), "DeleteBackendRequest")
	proto.RegisterType((*DeleteBackendResponse)(nil), "DeleteBackendResponse")
	proto.RegisterType((*BackendDetail)(nil), "BackendDetail")
	proto.RegisterType((*ListTypeRequest)(nil), "ListTypeRequest")
	proto.RegisterMapType((map[string]string)(nil), "ListTypeRequest.FilterEntry")
	proto.RegisterType((*ListTypeResponse)(nil), "ListTypeResponse")
	proto.RegisterType((*TypeDetail)(nil), "TypeDetail")
}

func init() { proto.RegisterFile("backend.proto", fileDescriptor_backend_89cca7ce8354284a) }

var fileDescriptor_backend_89cca7ce8354284a = []byte{
	// 630 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x55, 0x4d, 0x6f, 0xd3, 0x40,
	0x10, 0x8d, 0x93, 0x26, 0x4d, 0xc6, 0x6a, 0x69, 0xb7, 0x49, 0xb0, 0x2c, 0x04, 0xc1, 0x48, 0x28,
	0xe2, 0xb0, 0x12, 0x05, 0xa9, 0xd0, 0x03, 0x2a, 0xa5, 0x80, 0x2a, 0x10, 0x07, 0x8b, 0x5e, 0xb8,
	0xb9, 0xf6, 0x14, 0xad, 0x9a, 0xda, 0xc6, 0xbb, 0x41, 0xf8, 0xcc, 0x9f, 0xe5, 0xc4, 0x95, 0x2b,
	0xda, 0x0f, 0x27, 0xce, 0xc6, 0x08, 0x45, 0x42, 0xdc, 0xfc, 0x66, 0x66, 0x77, 0x66, 0xe7, 0xcd,
	0x3c, 0xc3, 0xce, 0x65, 0x14, 0x5f, 0x63, 0x9a, 0xd0, 0xbc, 0xc8, 0x44, 0x16, 0x9c, 0xc0, 0xf0,
	0x55, 0x81, 0x91, 0xc0, 0x53, 0x6d, 0x0e, 0xf1, 0xcb, 0x1c, 0xb9, 0x20, 0x53, 0xd8, 0x36, 0x81,
	0x9e, 0x33, 0x71, 0xa6, 0xee, 0xe1, 0x2e, 0x35, 0x11, 0x67, 0x28, 0x22, 0x36, 0x0b, 0x2b, 0x77,
	0xf0, 0x12, 0x46, 0xd6, 0x0d, 0x3c, 0xcf, 0x52, 0x8e, 0x1b, 0x5c, 0xf1, 0x00, 0xf6, 0xdf, 0xa2,
	0xb0, 0x2a, 0xd8, 0x85, 0x36, 0xd3, 0x27, 0x07, 0x61, 0x9b, 0x25, 0xc1, 0x0b, 0x20, 0xf5, 0xa0,
	0x8d, 0x93, 0xfc, 0x74, 0x80, 0xbc, 0x67, 0xdc, 0x4e, 0x33, 0x84, 0xee, 0x8c, 0xdd, 0x30, 0xa1,
	0x8e, 0x77, 0x43, 0x0d, 0xc8, 0x18, 0x7a, 0xd9, 0xd5, 0x15, 0x47, 0xe1, 0xb5, 0x95, 0xd9, 0x20,
	0xe2, 0x43, 0x9f, 0x67, 0x85, 0x78, 0x87, 0x25, 0xf7, 0x3a, 0x93, 0xce, 0x74, 0x10, 0x2e, 0x70,
	0xe5, 0x3b, 0x63, 0x05, 0xf7, 0xb6, 0x96, 0x3e, 0x89, 0xc9, 0x11, 0xf4, 0xde, 0xb0, 0x99, 0xc0,
	0xc2, 0xeb, 0x4e, 0x3a, 0x53, 0xf7, 0xf0, 0x1e, 0x5d, 0x2f, 0x85, 0xea, 0x88, 0xd7, 0xa9, 0x28,
	0xca, 0xd0, 0x84, 0xfb, 0xcf, 0xc1, 0xad, 0x99, 0xc9, 0x1e, 0x74, 0xae, 0xb1, 0x34, 0x5d, 0x91,
	0x9f, 0xb2, 0xfe, 0xaf, 0xd1, 0x6c, 0x8e, 0xaa, 0xd0, 0x41, 0xa8, 0xc1, 0x71, 0xfb, 0x99, 0x13,
	0x5c, 0xc0, 0xc1, 0x4a, 0x12, 0xd3, 0xb1, 0x47, 0xd0, 0x37, 0x2d, 0xe1, 0x9e, 0xa3, 0x8a, 0xb1,
	0x5b, 0xb6, 0xf0, 0x13, 0x02, 0x5b, 0x29, 0x7e, 0xab, 0x9a, 0xa0, 0xbe, 0x83, 0x4f, 0x30, 0xbc,
	0xc8, 0x93, 0xf5, 0x89, 0xb1, 0xf8, 0x92, 0x2d, 0x8c, 0xe2, 0x18, 0x39, 0x37, 0x95, 0x19, 0xa4,
	0xda, 0x84, 0xf1, 0xbc, 0x60, 0xa2, 0xf4, 0x3a, 0xca, 0xb3, 0xc0, 0x72, 0x96, 0xac, 0xbb, 0x37,
	0xa6, 0xf9, 0x21, 0x0c, 0xcf, 0x70, 0x86, 0x7f, 0x2b, 0x2f, 0xb8, 0x0d, 0x23, 0x2b, 0x4e, 0xa7,
	0x0a, 0xbe, 0xb7, 0x61, 0x67, 0xe5, 0xee, 0xb5, 0x97, 0xf9, 0xd0, 0x17, 0x98, 0x46, 0xa9, 0x38,
	0x4f, 0xcc, 0xdb, 0x16, 0x58, 0xbe, 0x7a, 0xce, 0xb1, 0x38, 0x4f, 0xcc, 0xdb, 0x0c, 0x52, 0x9d,
	0x8c, 0x6e, 0xd0, 0xdb, 0x52, 0x56, 0xf5, 0x2d, 0x6d, 0xa2, 0xcc, 0xd1, 0xeb, 0x6a, 0x9b, 0xfc,
	0x96, 0xe7, 0x0b, 0xfc, 0xcc, 0xb2, 0xd4, 0xeb, 0xe9, 0xf3, 0x1a, 0xc9, 0x9c, 0x98, 0x26, 0x79,
	0xc6, 0x52, 0xe1, 0x6d, 0xeb, 0x9c, 0x15, 0x26, 0x77, 0x01, 0x2e, 0xe7, 0xf1, 0x35, 0x8a, 0x0f,
	0x32, 0x43, 0x5f, 0x79, 0x6b, 0x96, 0x1a, 0x13, 0x83, 0x3f, 0x32, 0x01, 0x16, 0x13, 0x3f, 0x1c,
	0xb8, 0x25, 0xa7, 0xe7, 0x63, 0x99, 0xe3, 0xff, 0x5d, 0x95, 0xa7, 0xd6, 0xaa, 0xdc, 0xa1, 0x56,
	0x1d, 0xff, 0x7a, 0x4f, 0xce, 0x61, 0x6f, 0x99, 0xc1, 0xcc, 0xdb, 0x7d, 0xe8, 0x4a, 0x3a, 0xaa,
	0x0d, 0x71, 0xa9, 0xf4, 0x9a, 0x51, 0xd3, 0x9e, 0xc6, 0xdd, 0x38, 0x05, 0x58, 0x06, 0x2e, 0x38,
	0x77, 0x6a, 0x9c, 0x4f, 0xc0, 0x4d, 0x90, 0xc7, 0x05, 0xcb, 0x85, 0x24, 0x59, 0x17, 0x53, 0x37,
	0x1d, 0xfe, 0x6a, 0xc3, 0xb6, 0x99, 0x3f, 0x72, 0x02, 0x3b, 0x2b, 0xda, 0x4a, 0x46, 0xb4, 0x49,
	0xad, 0xfd, 0x31, 0x6d, 0x94, 0xe0, 0xa0, 0x45, 0x8e, 0x00, 0x96, 0xaa, 0x49, 0x08, 0x5d, 0xd3,
	0x59, 0xff, 0x80, 0xae, 0xcb, 0x6a, 0xd0, 0x22, 0xc7, 0xe0, 0xd6, 0xd4, 0x83, 0x1c, 0x34, 0x08,
	0x96, 0x3f, 0xa4, 0x0d, 0x02, 0x13, 0xb4, 0x64, 0xd9, 0x2b, 0x6b, 0x4c, 0x46, 0xb4, 0x49, 0x32,
	0xfc, 0x31, 0x6d, 0xdc, 0x76, 0x7d, 0xc3, 0xca, 0x76, 0x92, 0x11, 0x6d, 0xda, 0x6a, 0x7f, 0x4c,
	0x9b, 0x97, 0xb8, 0x45, 0x1e, 0x43, 0xbf, 0x62, 0x95, 0xec, 0xd9, 0x23, 0xe4, 0xef, 0x53, 0x9b,
	0xf2, 0xa0, 0x75, 0xd9, 0x53, 0xbf, 0xc4, 0x27, 0xbf, 0x03, 0x00, 0x00, 0xff, 0xff, 0x50, 0xb6,
	0xe0, 0x3d, 0x23, 0x07, 0x00, 0x00,
}
