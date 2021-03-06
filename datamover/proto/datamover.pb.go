// Code generated by protoc-gen-go. DO NOT EDIT.
// source: datamover.proto

package datamover

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

type KV struct {
	Key                  string   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value                string   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KV) Reset()         { *m = KV{} }
func (m *KV) String() string { return proto.CompactTextString(m) }
func (*KV) ProtoMessage()    {}
func (*KV) Descriptor() ([]byte, []int) {
	return fileDescriptor_datamover_f44375dc1e3daaa5, []int{0}
}
func (m *KV) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KV.Unmarshal(m, b)
}
func (m *KV) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KV.Marshal(b, m, deterministic)
}
func (dst *KV) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KV.Merge(dst, src)
}
func (m *KV) XXX_Size() int {
	return xxx_messageInfo_KV.Size(m)
}
func (m *KV) XXX_DiscardUnknown() {
	xxx_messageInfo_KV.DiscardUnknown(m)
}

var xxx_messageInfo_KV proto.InternalMessageInfo

func (m *KV) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *KV) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type Filter struct {
	Prefix               string   `protobuf:"bytes,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
	Tag                  []*KV    `protobuf:"bytes,2,rep,name=tag,proto3" json:"tag,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Filter) Reset()         { *m = Filter{} }
func (m *Filter) String() string { return proto.CompactTextString(m) }
func (*Filter) ProtoMessage()    {}
func (*Filter) Descriptor() ([]byte, []int) {
	return fileDescriptor_datamover_f44375dc1e3daaa5, []int{1}
}
func (m *Filter) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Filter.Unmarshal(m, b)
}
func (m *Filter) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Filter.Marshal(b, m, deterministic)
}
func (dst *Filter) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Filter.Merge(dst, src)
}
func (m *Filter) XXX_Size() int {
	return xxx_messageInfo_Filter.Size(m)
}
func (m *Filter) XXX_DiscardUnknown() {
	xxx_messageInfo_Filter.DiscardUnknown(m)
}

var xxx_messageInfo_Filter proto.InternalMessageInfo

func (m *Filter) GetPrefix() string {
	if m != nil {
		return m.Prefix
	}
	return ""
}

func (m *Filter) GetTag() []*KV {
	if m != nil {
		return m.Tag
	}
	return nil
}

type Connector struct {
	Type                 string   `protobuf:"bytes,1,opt,name=Type,proto3" json:"Type,omitempty"`
	BucketName           string   `protobuf:"bytes,2,opt,name=BucketName,proto3" json:"BucketName,omitempty"`
	ConnConfig           []*KV    `protobuf:"bytes,3,rep,name=ConnConfig,proto3" json:"ConnConfig,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Connector) Reset()         { *m = Connector{} }
func (m *Connector) String() string { return proto.CompactTextString(m) }
func (*Connector) ProtoMessage()    {}
func (*Connector) Descriptor() ([]byte, []int) {
	return fileDescriptor_datamover_f44375dc1e3daaa5, []int{2}
}
func (m *Connector) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Connector.Unmarshal(m, b)
}
func (m *Connector) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Connector.Marshal(b, m, deterministic)
}
func (dst *Connector) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Connector.Merge(dst, src)
}
func (m *Connector) XXX_Size() int {
	return xxx_messageInfo_Connector.Size(m)
}
func (m *Connector) XXX_DiscardUnknown() {
	xxx_messageInfo_Connector.DiscardUnknown(m)
}

var xxx_messageInfo_Connector proto.InternalMessageInfo

func (m *Connector) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *Connector) GetBucketName() string {
	if m != nil {
		return m.BucketName
	}
	return ""
}

func (m *Connector) GetConnConfig() []*KV {
	if m != nil {
		return m.ConnConfig
	}
	return nil
}

type AsistInfo struct {
	Type                 string   `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Details              []*KV    `protobuf:"bytes,2,rep,name=details,proto3" json:"details,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AsistInfo) Reset()         { *m = AsistInfo{} }
func (m *AsistInfo) String() string { return proto.CompactTextString(m) }
func (*AsistInfo) ProtoMessage()    {}
func (*AsistInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_datamover_f44375dc1e3daaa5, []int{3}
}
func (m *AsistInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AsistInfo.Unmarshal(m, b)
}
func (m *AsistInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AsistInfo.Marshal(b, m, deterministic)
}
func (dst *AsistInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AsistInfo.Merge(dst, src)
}
func (m *AsistInfo) XXX_Size() int {
	return xxx_messageInfo_AsistInfo.Size(m)
}
func (m *AsistInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_AsistInfo.DiscardUnknown(m)
}

var xxx_messageInfo_AsistInfo proto.InternalMessageInfo

func (m *AsistInfo) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *AsistInfo) GetDetails() []*KV {
	if m != nil {
		return m.Details
	}
	return nil
}

type RunJobRequest struct {
	Id                   string     `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	SourceConn           *Connector `protobuf:"bytes,2,opt,name=sourceConn,proto3" json:"sourceConn,omitempty"`
	DestConn             *Connector `protobuf:"bytes,3,opt,name=destConn,proto3" json:"destConn,omitempty"`
	Filt                 *Filter    `protobuf:"bytes,4,opt,name=filt,proto3" json:"filt,omitempty"`
	RemainSource         bool       `protobuf:"varint,5,opt,name=remainSource,proto3" json:"remainSource,omitempty"`
	Asist                *AsistInfo `protobuf:"bytes,6,opt,name=asist,proto3" json:"asist,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *RunJobRequest) Reset()         { *m = RunJobRequest{} }
func (m *RunJobRequest) String() string { return proto.CompactTextString(m) }
func (*RunJobRequest) ProtoMessage()    {}
func (*RunJobRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_datamover_f44375dc1e3daaa5, []int{4}
}
func (m *RunJobRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RunJobRequest.Unmarshal(m, b)
}
func (m *RunJobRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RunJobRequest.Marshal(b, m, deterministic)
}
func (dst *RunJobRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RunJobRequest.Merge(dst, src)
}
func (m *RunJobRequest) XXX_Size() int {
	return xxx_messageInfo_RunJobRequest.Size(m)
}
func (m *RunJobRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_RunJobRequest.DiscardUnknown(m)
}

var xxx_messageInfo_RunJobRequest proto.InternalMessageInfo

func (m *RunJobRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *RunJobRequest) GetSourceConn() *Connector {
	if m != nil {
		return m.SourceConn
	}
	return nil
}

func (m *RunJobRequest) GetDestConn() *Connector {
	if m != nil {
		return m.DestConn
	}
	return nil
}

func (m *RunJobRequest) GetFilt() *Filter {
	if m != nil {
		return m.Filt
	}
	return nil
}

func (m *RunJobRequest) GetRemainSource() bool {
	if m != nil {
		return m.RemainSource
	}
	return false
}

func (m *RunJobRequest) GetAsist() *AsistInfo {
	if m != nil {
		return m.Asist
	}
	return nil
}

type RunJobResponse struct {
	Err                  string   `protobuf:"bytes,1,opt,name=err,proto3" json:"err,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RunJobResponse) Reset()         { *m = RunJobResponse{} }
func (m *RunJobResponse) String() string { return proto.CompactTextString(m) }
func (*RunJobResponse) ProtoMessage()    {}
func (*RunJobResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_datamover_f44375dc1e3daaa5, []int{5}
}
func (m *RunJobResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RunJobResponse.Unmarshal(m, b)
}
func (m *RunJobResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RunJobResponse.Marshal(b, m, deterministic)
}
func (dst *RunJobResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RunJobResponse.Merge(dst, src)
}
func (m *RunJobResponse) XXX_Size() int {
	return xxx_messageInfo_RunJobResponse.Size(m)
}
func (m *RunJobResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_RunJobResponse.DiscardUnknown(m)
}

var xxx_messageInfo_RunJobResponse proto.InternalMessageInfo

func (m *RunJobResponse) GetErr() string {
	if m != nil {
		return m.Err
	}
	return ""
}

func init() {
	proto.RegisterType((*KV)(nil), "KV")
	proto.RegisterType((*Filter)(nil), "Filter")
	proto.RegisterType((*Connector)(nil), "Connector")
	proto.RegisterType((*AsistInfo)(nil), "AsistInfo")
	proto.RegisterType((*RunJobRequest)(nil), "RunJobRequest")
	proto.RegisterType((*RunJobResponse)(nil), "RunJobResponse")
}

func init() { proto.RegisterFile("datamover.proto", fileDescriptor_datamover_f44375dc1e3daaa5) }

var fileDescriptor_datamover_f44375dc1e3daaa5 = []byte{
	// 370 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x52, 0xdf, 0x8b, 0xd3, 0x40,
	0x10, 0x36, 0x49, 0x9b, 0xbb, 0xcc, 0x69, 0x4f, 0x06, 0x95, 0xa0, 0x28, 0x61, 0x05, 0x29, 0x2a,
	0x79, 0xa8, 0x0f, 0xfa, 0x24, 0xe8, 0x81, 0xa0, 0x07, 0x3e, 0xac, 0x72, 0xef, 0xdb, 0x66, 0x72,
	0xac, 0x4d, 0x77, 0xe3, 0xee, 0xa6, 0xd8, 0x7f, 0xd4, 0xbf, 0x47, 0x76, 0x9b, 0xc4, 0x94, 0x7b,
	0x9b, 0xf9, 0x66, 0xe6, 0xfb, 0xe6, 0x17, 0x5c, 0x56, 0xc2, 0x89, 0x9d, 0xde, 0x93, 0x29, 0x5b,
	0xa3, 0x9d, 0x66, 0x6f, 0x21, 0xbe, 0xbe, 0xc1, 0x87, 0x90, 0x6c, 0xe9, 0x90, 0x47, 0x45, 0xb4,
	0xcc, 0xb8, 0x37, 0xf1, 0x11, 0xcc, 0xf7, 0xa2, 0xe9, 0x28, 0x8f, 0x03, 0x76, 0x74, 0xd8, 0x7b,
	0x48, 0xbf, 0xc8, 0xc6, 0x91, 0xc1, 0x27, 0x90, 0xb6, 0x86, 0x6a, 0xf9, 0xa7, 0x2f, 0xea, 0x3d,
	0x7c, 0x0c, 0x89, 0x13, 0xb7, 0x79, 0x5c, 0x24, 0xcb, 0x8b, 0x55, 0x52, 0x5e, 0xdf, 0x70, 0xef,
	0xb3, 0x0a, 0xb2, 0x2b, 0xad, 0x14, 0x6d, 0x9c, 0x36, 0x88, 0x30, 0xfb, 0x79, 0x68, 0xa9, 0xaf,
	0x0c, 0x36, 0xbe, 0x00, 0xf8, 0xdc, 0x6d, 0xb6, 0xe4, 0xbe, 0x8b, 0xdd, 0x20, 0x3a, 0x41, 0xf0,
	0x25, 0x80, 0x27, 0xb8, 0xd2, 0xaa, 0x96, 0xb7, 0x79, 0xf2, 0x9f, 0x7e, 0x02, 0xb3, 0x8f, 0x90,
	0x7d, 0xb2, 0xd2, 0xba, 0xaf, 0xaa, 0xd6, 0x5e, 0xc5, 0x4d, 0x54, 0xbc, 0x8d, 0xcf, 0xe1, 0xac,
	0x22, 0x27, 0x64, 0x63, 0xa7, 0x1d, 0x0e, 0x18, 0xfb, 0x1b, 0xc1, 0x03, 0xde, 0xa9, 0x6f, 0x7a,
	0xcd, 0xe9, 0x77, 0x47, 0xd6, 0xe1, 0x02, 0x62, 0x59, 0xf5, 0x14, 0xb1, 0xac, 0xf0, 0x35, 0x80,
	0xd5, 0x9d, 0xd9, 0x90, 0x57, 0x0d, 0x6d, 0x5e, 0xac, 0xa0, 0x1c, 0x47, 0xe3, 0x93, 0x28, 0xbe,
	0x82, 0xf3, 0x8a, 0xac, 0x0b, 0x99, 0xc9, 0x9d, 0xcc, 0x31, 0x86, 0xcf, 0x60, 0x56, 0xcb, 0xc6,
	0xe5, 0xb3, 0x90, 0x73, 0x56, 0x1e, 0x37, 0xcc, 0x03, 0x88, 0x0c, 0xee, 0x1b, 0xda, 0x09, 0xa9,
	0x7e, 0x04, 0xe2, 0x7c, 0x5e, 0x44, 0xcb, 0x73, 0x7e, 0x82, 0x61, 0x01, 0x73, 0xe1, 0xc7, 0xce,
	0xd3, 0x5e, 0x65, 0x5c, 0x02, 0x3f, 0x06, 0x18, 0x83, 0xc5, 0x30, 0x97, 0x6d, 0xb5, 0xb2, 0xe4,
	0x2f, 0x4e, 0xc6, 0x0c, 0x17, 0x27, 0x63, 0x56, 0x1f, 0x20, 0x1b, 0x9f, 0x03, 0xdf, 0x40, 0xca,
	0x3b, 0xf5, 0x4b, 0xaf, 0x71, 0x51, 0x9e, 0x6c, 0xe4, 0xe9, 0x65, 0x79, 0xca, 0xc4, 0xee, 0xad,
	0xd3, 0xf0, 0x4a, 0xef, 0xfe, 0x05, 0x00, 0x00, 0xff, 0xff, 0x31, 0x83, 0x22, 0x6e, 0x5d, 0x02,
	0x00, 0x00,
}
