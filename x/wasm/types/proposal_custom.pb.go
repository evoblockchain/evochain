// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: x/wasm/proto/proposal_custom.proto

package types

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type UpdateDeploymentWhitelistProposal struct {
	Title                string   `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description          string   `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	DistributorAddrs     []string `protobuf:"bytes,3,rep,name=distributorAddrs,proto3" json:"distributorAddrs,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UpdateDeploymentWhitelistProposal) Reset()         { *m = UpdateDeploymentWhitelistProposal{} }
func (m *UpdateDeploymentWhitelistProposal) String() string { return proto.CompactTextString(m) }
func (*UpdateDeploymentWhitelistProposal) ProtoMessage()    {}
func (*UpdateDeploymentWhitelistProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_dd9d4d6e8a1d82c0, []int{0}
}
func (m *UpdateDeploymentWhitelistProposal) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateDeploymentWhitelistProposal.Unmarshal(m, b)
}
func (m *UpdateDeploymentWhitelistProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateDeploymentWhitelistProposal.Marshal(b, m, deterministic)
}
func (m *UpdateDeploymentWhitelistProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateDeploymentWhitelistProposal.Merge(m, src)
}
func (m *UpdateDeploymentWhitelistProposal) XXX_Size() int {
	return xxx_messageInfo_UpdateDeploymentWhitelistProposal.Size(m)
}
func (m *UpdateDeploymentWhitelistProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateDeploymentWhitelistProposal.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateDeploymentWhitelistProposal proto.InternalMessageInfo

func (m *UpdateDeploymentWhitelistProposal) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *UpdateDeploymentWhitelistProposal) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *UpdateDeploymentWhitelistProposal) GetDistributorAddrs() []string {
	if m != nil {
		return m.DistributorAddrs
	}
	return nil
}

type UpdateWASMContractMethodBlockedListProposal struct {
	Title                string           `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description          string           `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	BlockedMethods       *ContractMethods `protobuf:"bytes,3,opt,name=blockedMethods,proto3" json:"blockedMethods,omitempty"`
	IsDelete             bool             `protobuf:"varint,4,opt,name=isDelete,proto3" json:"isDelete,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *UpdateWASMContractMethodBlockedListProposal) Reset() {
	*m = UpdateWASMContractMethodBlockedListProposal{}
}
func (m *UpdateWASMContractMethodBlockedListProposal) String() string {
	return proto.CompactTextString(m)
}
func (*UpdateWASMContractMethodBlockedListProposal) ProtoMessage() {}
func (*UpdateWASMContractMethodBlockedListProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_dd9d4d6e8a1d82c0, []int{1}
}
func (m *UpdateWASMContractMethodBlockedListProposal) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateWASMContractMethodBlockedListProposal.Unmarshal(m, b)
}
func (m *UpdateWASMContractMethodBlockedListProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateWASMContractMethodBlockedListProposal.Marshal(b, m, deterministic)
}
func (m *UpdateWASMContractMethodBlockedListProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateWASMContractMethodBlockedListProposal.Merge(m, src)
}
func (m *UpdateWASMContractMethodBlockedListProposal) XXX_Size() int {
	return xxx_messageInfo_UpdateWASMContractMethodBlockedListProposal.Size(m)
}
func (m *UpdateWASMContractMethodBlockedListProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateWASMContractMethodBlockedListProposal.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateWASMContractMethodBlockedListProposal proto.InternalMessageInfo

func (m *UpdateWASMContractMethodBlockedListProposal) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *UpdateWASMContractMethodBlockedListProposal) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *UpdateWASMContractMethodBlockedListProposal) GetBlockedMethods() *ContractMethods {
	if m != nil {
		return m.BlockedMethods
	}
	return nil
}

func (m *UpdateWASMContractMethodBlockedListProposal) GetIsDelete() bool {
	if m != nil {
		return m.IsDelete
	}
	return false
}

type ContractMethods struct {
	ContractAddr         string    `protobuf:"bytes,1,opt,name=contractAddr,proto3" json:"contractAddr,omitempty"`
	Methods              []*Method `protobuf:"bytes,2,rep,name=methods,proto3" json:"methods,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *ContractMethods) Reset()         { *m = ContractMethods{} }
func (m *ContractMethods) String() string { return proto.CompactTextString(m) }
func (*ContractMethods) ProtoMessage()    {}
func (*ContractMethods) Descriptor() ([]byte, []int) {
	return fileDescriptor_dd9d4d6e8a1d82c0, []int{2}
}
func (m *ContractMethods) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ContractMethods.Unmarshal(m, b)
}
func (m *ContractMethods) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ContractMethods.Marshal(b, m, deterministic)
}
func (m *ContractMethods) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ContractMethods.Merge(m, src)
}
func (m *ContractMethods) XXX_Size() int {
	return xxx_messageInfo_ContractMethods.Size(m)
}
func (m *ContractMethods) XXX_DiscardUnknown() {
	xxx_messageInfo_ContractMethods.DiscardUnknown(m)
}

var xxx_messageInfo_ContractMethods proto.InternalMessageInfo

func (m *ContractMethods) GetContractAddr() string {
	if m != nil {
		return m.ContractAddr
	}
	return ""
}

func (m *ContractMethods) GetMethods() []*Method {
	if m != nil {
		return m.Methods
	}
	return nil
}

type Method struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Extra                string   `protobuf:"bytes,2,opt,name=extra,proto3" json:"extra,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Method) Reset()         { *m = Method{} }
func (m *Method) String() string { return proto.CompactTextString(m) }
func (*Method) ProtoMessage()    {}
func (*Method) Descriptor() ([]byte, []int) {
	return fileDescriptor_dd9d4d6e8a1d82c0, []int{3}
}
func (m *Method) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Method.Unmarshal(m, b)
}
func (m *Method) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Method.Marshal(b, m, deterministic)
}
func (m *Method) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Method.Merge(m, src)
}
func (m *Method) XXX_Size() int {
	return xxx_messageInfo_Method.Size(m)
}
func (m *Method) XXX_DiscardUnknown() {
	xxx_messageInfo_Method.DiscardUnknown(m)
}

var xxx_messageInfo_Method proto.InternalMessageInfo

func (m *Method) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Method) GetExtra() string {
	if m != nil {
		return m.Extra
	}
	return ""
}

func init() {
	proto.RegisterType((*UpdateDeploymentWhitelistProposal)(nil), "types.UpdateDeploymentWhitelistProposal")
	proto.RegisterType((*UpdateWASMContractMethodBlockedListProposal)(nil), "types.UpdateWASMContractMethodBlockedListProposal")
	proto.RegisterType((*ContractMethods)(nil), "types.ContractMethods")
	proto.RegisterType((*Method)(nil), "types.Method")
}

func init() {
	proto.RegisterFile("x/wasm/proto/proposal_custom.proto", fileDescriptor_dd9d4d6e8a1d82c0)
}

var fileDescriptor_dd9d4d6e8a1d82c0 = []byte{
	// 315 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x92, 0xcf, 0x4a, 0x03, 0x31,
	0x10, 0xc6, 0xd9, 0xfe, 0xb3, 0x9d, 0xd6, 0x2a, 0x41, 0x64, 0xf1, 0xb4, 0xee, 0xc5, 0x45, 0xa1,
	0x85, 0x7a, 0x17, 0x5a, 0x7b, 0xb4, 0x20, 0x2b, 0x52, 0xf0, 0xa0, 0xec, 0x6e, 0x06, 0x1a, 0xdc,
	0xdd, 0x2c, 0xc9, 0x14, 0xdb, 0x27, 0xf0, 0xb9, 0x7c, 0x33, 0xd9, 0x24, 0x15, 0x5b, 0xaf, 0x5e,
	0x42, 0xe6, 0x9b, 0xcc, 0x97, 0xdf, 0x4c, 0x02, 0xe1, 0x66, 0xfc, 0x91, 0xe8, 0x62, 0x5c, 0x29,
	0x49, 0xb2, 0x5e, 0x2b, 0xa9, 0x93, 0xfc, 0x2d, 0x5b, 0x6b, 0x92, 0xc5, 0xc8, 0xa8, 0xac, 0x4d,
	0xdb, 0x0a, 0x75, 0xf8, 0xe9, 0xc1, 0xe5, 0x73, 0xc5, 0x13, 0xc2, 0x39, 0x56, 0xb9, 0xdc, 0x16,
	0x58, 0xd2, 0x72, 0x25, 0x08, 0x73, 0xa1, 0xe9, 0xd1, 0x55, 0xb2, 0x33, 0x68, 0x93, 0xa0, 0x1c,
	0x7d, 0x2f, 0xf0, 0xa2, 0x5e, 0x6c, 0x03, 0x16, 0x40, 0x9f, 0xa3, 0xce, 0x94, 0xa8, 0x48, 0xc8,
	0xd2, 0x6f, 0x98, 0xdc, 0x6f, 0x89, 0x5d, 0xc3, 0x29, 0x17, 0x9a, 0x94, 0x48, 0xd7, 0x24, 0xd5,
	0x94, 0x73, 0xa5, 0xfd, 0x66, 0xd0, 0x8c, 0x7a, 0xf1, 0x1f, 0x3d, 0xfc, 0xf2, 0xe0, 0xc6, 0x92,
	0x2c, 0xa7, 0x4f, 0x8b, 0x7b, 0x59, 0x92, 0x4a, 0x32, 0x5a, 0x20, 0xad, 0x24, 0x9f, 0xe5, 0x32,
	0x7b, 0x47, 0xfe, 0xf0, 0x1f, 0x4c, 0x77, 0x30, 0x4c, 0xad, 0x9d, 0xf5, 0xae, 0x89, 0xbc, 0xa8,
	0x3f, 0x39, 0x1f, 0x99, 0x89, 0x8c, 0xf6, 0x6f, 0xd6, 0xf1, 0xc1, 0x69, 0x76, 0x01, 0x5d, 0xa1,
	0xe7, 0x98, 0x23, 0xa1, 0xdf, 0x0a, 0xbc, 0xa8, 0x1b, 0xff, 0xc4, 0xe1, 0x2b, 0x9c, 0x1c, 0x94,
	0xb3, 0x10, 0x06, 0x99, 0x93, 0xea, 0x3e, 0x1d, 0xed, 0x9e, 0xc6, 0xae, 0xe0, 0xa8, 0x70, 0x2c,
	0x8d, 0xa0, 0x19, 0xf5, 0x27, 0xc7, 0x8e, 0xc5, 0x9a, 0xc4, 0xbb, 0x6c, 0x38, 0x81, 0x8e, 0x95,
	0x18, 0x83, 0x56, 0x99, 0x14, 0xbb, 0xe6, 0xcd, 0xbe, 0x9e, 0x08, 0x6e, 0x48, 0x25, 0xae, 0x6b,
	0x1b, 0xcc, 0x86, 0x2f, 0x03, 0xf7, 0x1d, 0x8c, 0x67, 0xda, 0x31, 0xef, 0x7f, 0xfb, 0x1d, 0x00,
	0x00, 0xff, 0xff, 0x73, 0x2c, 0x40, 0xd8, 0x25, 0x02, 0x00, 0x00,
}