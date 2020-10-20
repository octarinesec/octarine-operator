// Code generated by protoc-gen-go. DO NOT EDIT.
// source: monitor.proto

package monitor

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ServiceHealthReport_Kind int32

const (
	ServiceHealthReport_DEPLOYMENT ServiceHealthReport_Kind = 0
	ServiceHealthReport_DAEMONSET  ServiceHealthReport_Kind = 1
)

var ServiceHealthReport_Kind_name = map[int32]string{
	0: "DEPLOYMENT",
	1: "DAEMONSET",
}

var ServiceHealthReport_Kind_value = map[string]int32{
	"DEPLOYMENT": 0,
	"DAEMONSET":  1,
}

func (x ServiceHealthReport_Kind) String() string {
	return proto.EnumName(ServiceHealthReport_Kind_name, int32(x))
}

func (ServiceHealthReport_Kind) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_44174b7b2a306b71, []int{4, 0}
}

type WebhookHealthReport_WebhookType int32

const (
	WebhookHealthReport_VALIDATING WebhookHealthReport_WebhookType = 0
	WebhookHealthReport_MUTATING   WebhookHealthReport_WebhookType = 1
)

var WebhookHealthReport_WebhookType_name = map[int32]string{
	0: "VALIDATING",
	1: "MUTATING",
}

var WebhookHealthReport_WebhookType_value = map[string]int32{
	"VALIDATING": 0,
	"MUTATING":   1,
}

func (x WebhookHealthReport_WebhookType) String() string {
	return proto.EnumName(WebhookHealthReport_WebhookType_name, int32(x))
}

func (WebhookHealthReport_WebhookType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_44174b7b2a306b71, []int{5, 0}
}

type ContainerSpec struct {
	Image                string   `protobuf:"bytes,1,opt,name=image,proto3" json:"image,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ContainerSpec) Reset()         { *m = ContainerSpec{} }
func (m *ContainerSpec) String() string { return proto.CompactTextString(m) }
func (*ContainerSpec) ProtoMessage()    {}
func (*ContainerSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_44174b7b2a306b71, []int{0}
}

func (m *ContainerSpec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ContainerSpec.Unmarshal(m, b)
}
func (m *ContainerSpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ContainerSpec.Marshal(b, m, deterministic)
}
func (m *ContainerSpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ContainerSpec.Merge(m, src)
}
func (m *ContainerSpec) XXX_Size() int {
	return xxx_messageInfo_ContainerSpec.Size(m)
}
func (m *ContainerSpec) XXX_DiscardUnknown() {
	xxx_messageInfo_ContainerSpec.DiscardUnknown(m)
}

var xxx_messageInfo_ContainerSpec proto.InternalMessageInfo

func (m *ContainerSpec) GetImage() string {
	if m != nil {
		return m.Image
	}
	return ""
}

type ReplicaSpec struct {
	Containers           map[string]*ContainerSpec `protobuf:"bytes,1,rep,name=containers,proto3" json:"containers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *ReplicaSpec) Reset()         { *m = ReplicaSpec{} }
func (m *ReplicaSpec) String() string { return proto.CompactTextString(m) }
func (*ReplicaSpec) ProtoMessage()    {}
func (*ReplicaSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_44174b7b2a306b71, []int{1}
}

func (m *ReplicaSpec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReplicaSpec.Unmarshal(m, b)
}
func (m *ReplicaSpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReplicaSpec.Marshal(b, m, deterministic)
}
func (m *ReplicaSpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReplicaSpec.Merge(m, src)
}
func (m *ReplicaSpec) XXX_Size() int {
	return xxx_messageInfo_ReplicaSpec.Size(m)
}
func (m *ReplicaSpec) XXX_DiscardUnknown() {
	xxx_messageInfo_ReplicaSpec.DiscardUnknown(m)
}

var xxx_messageInfo_ReplicaSpec proto.InternalMessageInfo

func (m *ReplicaSpec) GetContainers() map[string]*ContainerSpec {
	if m != nil {
		return m.Containers
	}
	return nil
}

type ReplicaHealth struct {
	Node                 string       `protobuf:"bytes,1,opt,name=node,proto3" json:"node,omitempty"`
	Spec                 *ReplicaSpec `protobuf:"bytes,2,opt,name=spec,proto3" json:"spec,omitempty"`
	Status               []byte       `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *ReplicaHealth) Reset()         { *m = ReplicaHealth{} }
func (m *ReplicaHealth) String() string { return proto.CompactTextString(m) }
func (*ReplicaHealth) ProtoMessage()    {}
func (*ReplicaHealth) Descriptor() ([]byte, []int) {
	return fileDescriptor_44174b7b2a306b71, []int{2}
}

func (m *ReplicaHealth) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReplicaHealth.Unmarshal(m, b)
}
func (m *ReplicaHealth) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReplicaHealth.Marshal(b, m, deterministic)
}
func (m *ReplicaHealth) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReplicaHealth.Merge(m, src)
}
func (m *ReplicaHealth) XXX_Size() int {
	return xxx_messageInfo_ReplicaHealth.Size(m)
}
func (m *ReplicaHealth) XXX_DiscardUnknown() {
	xxx_messageInfo_ReplicaHealth.DiscardUnknown(m)
}

var xxx_messageInfo_ReplicaHealth proto.InternalMessageInfo

func (m *ReplicaHealth) GetNode() string {
	if m != nil {
		return m.Node
	}
	return ""
}

func (m *ReplicaHealth) GetSpec() *ReplicaSpec {
	if m != nil {
		return m.Spec
	}
	return nil
}

func (m *ReplicaHealth) GetStatus() []byte {
	if m != nil {
		return m.Status
	}
	return nil
}

type ServiceSpec struct {
	Replicas             int32                     `protobuf:"varint,1,opt,name=replicas,proto3" json:"replicas,omitempty"`
	Containers           map[string]*ContainerSpec `protobuf:"bytes,2,rep,name=containers,proto3" json:"containers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *ServiceSpec) Reset()         { *m = ServiceSpec{} }
func (m *ServiceSpec) String() string { return proto.CompactTextString(m) }
func (*ServiceSpec) ProtoMessage()    {}
func (*ServiceSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_44174b7b2a306b71, []int{3}
}

func (m *ServiceSpec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServiceSpec.Unmarshal(m, b)
}
func (m *ServiceSpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServiceSpec.Marshal(b, m, deterministic)
}
func (m *ServiceSpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServiceSpec.Merge(m, src)
}
func (m *ServiceSpec) XXX_Size() int {
	return xxx_messageInfo_ServiceSpec.Size(m)
}
func (m *ServiceSpec) XXX_DiscardUnknown() {
	xxx_messageInfo_ServiceSpec.DiscardUnknown(m)
}

var xxx_messageInfo_ServiceSpec proto.InternalMessageInfo

func (m *ServiceSpec) GetReplicas() int32 {
	if m != nil {
		return m.Replicas
	}
	return 0
}

func (m *ServiceSpec) GetContainers() map[string]*ContainerSpec {
	if m != nil {
		return m.Containers
	}
	return nil
}

type ServiceHealthReport struct {
	Kind                 ServiceHealthReport_Kind  `protobuf:"varint,1,opt,name=kind,proto3,enum=monitor.ServiceHealthReport_Kind" json:"kind,omitempty"`
	Spec                 *ServiceSpec              `protobuf:"bytes,2,opt,name=spec,proto3" json:"spec,omitempty"`
	Status               []byte                    `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	Replicas             map[string]*ReplicaHealth `protobuf:"bytes,4,rep,name=replicas,proto3" json:"replicas,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *ServiceHealthReport) Reset()         { *m = ServiceHealthReport{} }
func (m *ServiceHealthReport) String() string { return proto.CompactTextString(m) }
func (*ServiceHealthReport) ProtoMessage()    {}
func (*ServiceHealthReport) Descriptor() ([]byte, []int) {
	return fileDescriptor_44174b7b2a306b71, []int{4}
}

func (m *ServiceHealthReport) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServiceHealthReport.Unmarshal(m, b)
}
func (m *ServiceHealthReport) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServiceHealthReport.Marshal(b, m, deterministic)
}
func (m *ServiceHealthReport) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServiceHealthReport.Merge(m, src)
}
func (m *ServiceHealthReport) XXX_Size() int {
	return xxx_messageInfo_ServiceHealthReport.Size(m)
}
func (m *ServiceHealthReport) XXX_DiscardUnknown() {
	xxx_messageInfo_ServiceHealthReport.DiscardUnknown(m)
}

var xxx_messageInfo_ServiceHealthReport proto.InternalMessageInfo

func (m *ServiceHealthReport) GetKind() ServiceHealthReport_Kind {
	if m != nil {
		return m.Kind
	}
	return ServiceHealthReport_DEPLOYMENT
}

func (m *ServiceHealthReport) GetSpec() *ServiceSpec {
	if m != nil {
		return m.Spec
	}
	return nil
}

func (m *ServiceHealthReport) GetStatus() []byte {
	if m != nil {
		return m.Status
	}
	return nil
}

func (m *ServiceHealthReport) GetReplicas() map[string]*ReplicaHealth {
	if m != nil {
		return m.Replicas
	}
	return nil
}

type WebhookHealthReport struct {
	WebhookType          WebhookHealthReport_WebhookType `protobuf:"varint,1,opt,name=webhookType,proto3,enum=monitor.WebhookHealthReport_WebhookType" json:"webhookType,omitempty"`
	Uid                  string                          `protobuf:"bytes,2,opt,name=uid,proto3" json:"uid,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                        `json:"-"`
	XXX_unrecognized     []byte                          `json:"-"`
	XXX_sizecache        int32                           `json:"-"`
}

func (m *WebhookHealthReport) Reset()         { *m = WebhookHealthReport{} }
func (m *WebhookHealthReport) String() string { return proto.CompactTextString(m) }
func (*WebhookHealthReport) ProtoMessage()    {}
func (*WebhookHealthReport) Descriptor() ([]byte, []int) {
	return fileDescriptor_44174b7b2a306b71, []int{5}
}

func (m *WebhookHealthReport) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WebhookHealthReport.Unmarshal(m, b)
}
func (m *WebhookHealthReport) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WebhookHealthReport.Marshal(b, m, deterministic)
}
func (m *WebhookHealthReport) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WebhookHealthReport.Merge(m, src)
}
func (m *WebhookHealthReport) XXX_Size() int {
	return xxx_messageInfo_WebhookHealthReport.Size(m)
}
func (m *WebhookHealthReport) XXX_DiscardUnknown() {
	xxx_messageInfo_WebhookHealthReport.DiscardUnknown(m)
}

var xxx_messageInfo_WebhookHealthReport proto.InternalMessageInfo

func (m *WebhookHealthReport) GetWebhookType() WebhookHealthReport_WebhookType {
	if m != nil {
		return m.WebhookType
	}
	return WebhookHealthReport_VALIDATING
}

func (m *WebhookHealthReport) GetUid() string {
	if m != nil {
		return m.Uid
	}
	return ""
}

type HealthReport struct {
	Account              string                          `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	Domain               string                          `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	Webhooks             map[string]*WebhookHealthReport `protobuf:"bytes,3,rep,name=webhooks,proto3" json:"webhooks,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Services             map[string]*ServiceHealthReport `protobuf:"bytes,4,rep,name=services,proto3" json:"services,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	EnableComponents     map[string]bool                 `protobuf:"bytes,5,rep,name=enableComponents,proto3" json:"enableComponents,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	Version              string                          `protobuf:"bytes,6,opt,name=version,proto3" json:"version,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                        `json:"-"`
	XXX_unrecognized     []byte                          `json:"-"`
	XXX_sizecache        int32                           `json:"-"`
}

func (m *HealthReport) Reset()         { *m = HealthReport{} }
func (m *HealthReport) String() string { return proto.CompactTextString(m) }
func (*HealthReport) ProtoMessage()    {}
func (*HealthReport) Descriptor() ([]byte, []int) {
	return fileDescriptor_44174b7b2a306b71, []int{6}
}

func (m *HealthReport) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HealthReport.Unmarshal(m, b)
}
func (m *HealthReport) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HealthReport.Marshal(b, m, deterministic)
}
func (m *HealthReport) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HealthReport.Merge(m, src)
}
func (m *HealthReport) XXX_Size() int {
	return xxx_messageInfo_HealthReport.Size(m)
}
func (m *HealthReport) XXX_DiscardUnknown() {
	xxx_messageInfo_HealthReport.DiscardUnknown(m)
}

var xxx_messageInfo_HealthReport proto.InternalMessageInfo

func (m *HealthReport) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

func (m *HealthReport) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	return ""
}

func (m *HealthReport) GetWebhooks() map[string]*WebhookHealthReport {
	if m != nil {
		return m.Webhooks
	}
	return nil
}

func (m *HealthReport) GetServices() map[string]*ServiceHealthReport {
	if m != nil {
		return m.Services
	}
	return nil
}

func (m *HealthReport) GetEnableComponents() map[string]bool {
	if m != nil {
		return m.EnableComponents
	}
	return nil
}

func (m *HealthReport) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

type HealthReportReply struct {
	Ack                  bool     `protobuf:"varint,1,opt,name=ack,proto3" json:"ack,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HealthReportReply) Reset()         { *m = HealthReportReply{} }
func (m *HealthReportReply) String() string { return proto.CompactTextString(m) }
func (*HealthReportReply) ProtoMessage()    {}
func (*HealthReportReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_44174b7b2a306b71, []int{7}
}

func (m *HealthReportReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HealthReportReply.Unmarshal(m, b)
}
func (m *HealthReportReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HealthReportReply.Marshal(b, m, deterministic)
}
func (m *HealthReportReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HealthReportReply.Merge(m, src)
}
func (m *HealthReportReply) XXX_Size() int {
	return xxx_messageInfo_HealthReportReply.Size(m)
}
func (m *HealthReportReply) XXX_DiscardUnknown() {
	xxx_messageInfo_HealthReportReply.DiscardUnknown(m)
}

var xxx_messageInfo_HealthReportReply proto.InternalMessageInfo

func (m *HealthReportReply) GetAck() bool {
	if m != nil {
		return m.Ack
	}
	return false
}

func init() {
	proto.RegisterEnum("monitor.ServiceHealthReport_Kind", ServiceHealthReport_Kind_name, ServiceHealthReport_Kind_value)
	proto.RegisterEnum("monitor.WebhookHealthReport_WebhookType", WebhookHealthReport_WebhookType_name, WebhookHealthReport_WebhookType_value)
	proto.RegisterType((*ContainerSpec)(nil), "monitor.ContainerSpec")
	proto.RegisterType((*ReplicaSpec)(nil), "monitor.ReplicaSpec")
	proto.RegisterMapType((map[string]*ContainerSpec)(nil), "monitor.ReplicaSpec.ContainersEntry")
	proto.RegisterType((*ReplicaHealth)(nil), "monitor.ReplicaHealth")
	proto.RegisterType((*ServiceSpec)(nil), "monitor.ServiceSpec")
	proto.RegisterMapType((map[string]*ContainerSpec)(nil), "monitor.ServiceSpec.ContainersEntry")
	proto.RegisterType((*ServiceHealthReport)(nil), "monitor.ServiceHealthReport")
	proto.RegisterMapType((map[string]*ReplicaHealth)(nil), "monitor.ServiceHealthReport.ReplicasEntry")
	proto.RegisterType((*WebhookHealthReport)(nil), "monitor.WebhookHealthReport")
	proto.RegisterType((*HealthReport)(nil), "monitor.HealthReport")
	proto.RegisterMapType((map[string]bool)(nil), "monitor.HealthReport.EnableComponentsEntry")
	proto.RegisterMapType((map[string]*ServiceHealthReport)(nil), "monitor.HealthReport.ServicesEntry")
	proto.RegisterMapType((map[string]*WebhookHealthReport)(nil), "monitor.HealthReport.WebhooksEntry")
	proto.RegisterType((*HealthReportReply)(nil), "monitor.HealthReportReply")
}

func init() { proto.RegisterFile("monitor.proto", fileDescriptor_44174b7b2a306b71) }

var fileDescriptor_44174b7b2a306b71 = []byte{
	// 638 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xc4, 0x55, 0xcd, 0x6e, 0xd3, 0x4c,
	0x14, 0xad, 0x53, 0xa7, 0x4d, 0xaf, 0x9b, 0x7e, 0xf9, 0xa6, 0x3f, 0xb2, 0x2c, 0x16, 0xc1, 0x50,
	0xc9, 0xa2, 0x28, 0x8b, 0x20, 0x24, 0xc4, 0x06, 0x55, 0x8d, 0xa1, 0x85, 0xa6, 0x45, 0x13, 0x97,
	0xaa, 0x4b, 0xd7, 0x1e, 0x51, 0x2b, 0xce, 0x8c, 0x65, 0x3b, 0x45, 0x79, 0x1b, 0xd6, 0xac, 0x78,
	0x0a, 0x1e, 0x80, 0x27, 0x42, 0x33, 0x1e, 0x5b, 0xe3, 0xe0, 0x46, 0xea, 0x8a, 0xdd, 0xdc, 0xc9,
	0xbd, 0xe7, 0x9c, 0x7b, 0xce, 0x24, 0x81, 0xee, 0x8c, 0xd1, 0x28, 0x67, 0xe9, 0x20, 0x49, 0x59,
	0xce, 0xd0, 0xa6, 0x2c, 0xed, 0x43, 0xe8, 0x9e, 0x30, 0x9a, 0xfb, 0x11, 0x25, 0xe9, 0x24, 0x21,
	0x01, 0xda, 0x83, 0x76, 0x34, 0xf3, 0xbf, 0x12, 0x53, 0xeb, 0x6b, 0xce, 0x16, 0x2e, 0x0a, 0xfb,
	0x87, 0x06, 0x06, 0x26, 0x49, 0x1c, 0x05, 0xbe, 0xe8, 0x1a, 0x01, 0x04, 0xe5, 0x58, 0x66, 0x6a,
	0xfd, 0x75, 0xc7, 0x18, 0x3e, 0x1f, 0x94, 0x1c, 0x4a, 0xe7, 0xa0, 0x42, 0xcf, 0x5c, 0x9a, 0xa7,
	0x0b, 0xac, 0xcc, 0x59, 0x57, 0xf0, 0xdf, 0xd2, 0xc7, 0xa8, 0x07, 0xeb, 0x53, 0xb2, 0x90, 0xe4,
	0xfc, 0x88, 0x5e, 0x42, 0xfb, 0xde, 0x8f, 0xe7, 0xc4, 0x6c, 0xf5, 0x35, 0xc7, 0x18, 0x1e, 0x54,
	0x2c, 0x35, 0xdd, 0xb8, 0x68, 0x7a, 0xdb, 0x7a, 0xa3, 0xd9, 0x04, 0xba, 0x52, 0xc1, 0x29, 0xf1,
	0xe3, 0xfc, 0x0e, 0x21, 0xd0, 0x29, 0x0b, 0xcb, 0x95, 0xc4, 0x19, 0x39, 0xa0, 0x67, 0x09, 0x09,
	0x24, 0xea, 0x5e, 0x93, 0x76, 0x2c, 0x3a, 0xd0, 0x01, 0x6c, 0x64, 0xb9, 0x9f, 0xcf, 0x33, 0x73,
	0xbd, 0xaf, 0x39, 0xdb, 0x58, 0x56, 0xf6, 0x2f, 0x0d, 0x8c, 0x09, 0x49, 0xef, 0xa3, 0x80, 0x08,
	0x4f, 0x2c, 0xe8, 0xa4, 0xc5, 0x70, 0x26, 0x98, 0xda, 0xb8, 0xaa, 0x97, 0xfc, 0x6a, 0x2d, 0xf9,
	0xa5, 0xa0, 0xfc, 0x0b, 0xbf, 0x7e, 0xb7, 0x60, 0x57, 0x4a, 0x28, 0x0c, 0xc3, 0x24, 0x61, 0x69,
	0x8e, 0x5e, 0x83, 0x3e, 0x8d, 0x68, 0x28, 0xc0, 0x77, 0x86, 0x4f, 0x97, 0xe5, 0xaa, 0xbd, 0x83,
	0x4f, 0x11, 0x0d, 0xb1, 0x68, 0x7f, 0xd0, 0x59, 0x65, 0xcb, 0xd5, 0xce, 0xa2, 0xf7, 0x8a, 0x93,
	0xba, 0xf0, 0xea, 0xc5, 0x4a, 0x72, 0x99, 0x99, 0x74, 0xac, 0x9a, 0xb5, 0x26, 0xd5, 0x43, 0x78,
	0xbc, 0x5b, 0xb5, 0x17, 0xa4, 0xba, 0x75, 0x08, 0x3a, 0x5f, 0x16, 0xed, 0x00, 0x8c, 0xdc, 0xcf,
	0xe7, 0x97, 0x37, 0x63, 0xf7, 0xc2, 0xeb, 0xad, 0xa1, 0x2e, 0x6c, 0x8d, 0x8e, 0xdd, 0xf1, 0xe5,
	0xc5, 0xc4, 0xf5, 0x7a, 0x9a, 0xfd, 0x5d, 0x83, 0xdd, 0x6b, 0x72, 0x7b, 0xc7, 0xd8, 0xb4, 0x66,
	0xea, 0x47, 0x30, 0xbe, 0x15, 0xd7, 0xde, 0x22, 0x21, 0xd2, 0x5b, 0xa7, 0xa2, 0x6d, 0x18, 0x29,
	0xef, 0x78, 0x3f, 0x56, 0x87, 0xf9, 0x3a, 0xf3, 0x28, 0x14, 0xd2, 0xb7, 0x30, 0x3f, 0xda, 0x47,
	0x60, 0x28, 0xdd, 0x5c, 0xe3, 0x97, 0xe3, 0xf3, 0xb3, 0xd1, 0xb1, 0x77, 0x76, 0xf1, 0xa1, 0xb7,
	0x86, 0xb6, 0xa1, 0x33, 0xbe, 0xf2, 0x8a, 0x4a, 0xb3, 0x7f, 0xea, 0xb0, 0x5d, 0xd3, 0x66, 0xc2,
	0xa6, 0x1f, 0x04, 0x6c, 0x4e, 0x73, 0x69, 0x51, 0x59, 0xf2, 0xa4, 0x42, 0x36, 0xf3, 0x23, 0x2a,
	0xc9, 0x64, 0x85, 0xde, 0x41, 0x47, 0x0a, 0xe2, 0x19, 0xf2, 0xa4, 0x9e, 0x55, 0xab, 0x34, 0xed,
	0x50, 0x46, 0x54, 0x0e, 0x71, 0x80, 0xac, 0x48, 0xb4, 0x8c, 0xfa, 0x01, 0x00, 0x99, 0x7b, 0x09,
	0x50, 0x0e, 0xa1, 0x6b, 0xe8, 0x11, 0xea, 0xdf, 0xc6, 0xe4, 0x84, 0xcd, 0x12, 0x46, 0x09, 0xcd,
	0x33, 0xb3, 0x2d, 0x80, 0x8e, 0x9a, 0x81, 0xdc, 0xa5, 0xee, 0x02, 0xf0, 0x2f, 0x10, 0x6e, 0xc6,
	0x3d, 0x49, 0xb3, 0x88, 0x51, 0x73, 0xa3, 0x30, 0x43, 0x96, 0xd6, 0x0d, 0x74, 0x6b, 0xeb, 0x34,
	0x3c, 0xab, 0x61, 0xfd, 0x59, 0x3d, 0x59, 0x95, 0xaf, 0xf2, 0xb8, 0x38, 0x74, 0x6d, 0xd1, 0xc7,
	0x40, 0x37, 0x7c, 0x33, 0x54, 0xe8, 0x13, 0xd8, 0x6f, 0x5c, 0xbd, 0x81, 0x62, 0x4f, 0xa5, 0xe8,
	0xd4, 0x1f, 0xff, 0xff, 0x35, 0x7c, 0x92, 0xc4, 0x02, 0xc0, 0x0f, 0xa6, 0x02, 0xa0, 0x83, 0xf9,
	0x71, 0xe8, 0xc1, 0xe6, 0xb8, 0x50, 0x85, 0xce, 0x00, 0x9d, 0xfa, 0x34, 0x8c, 0xeb, 0x3f, 0x2d,
	0xfb, 0x8d, 0xd9, 0x58, 0x56, 0xe3, 0xb5, 0x60, 0xb1, 0xd7, 0x6e, 0x37, 0xc4, 0x7f, 0xd7, 0xab,
	0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x9b, 0x40, 0x0d, 0x98, 0xcc, 0x06, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MonitorClient is the client API for Monitor service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MonitorClient interface {
	HandleHealthReport(ctx context.Context, in *HealthReport, opts ...grpc.CallOption) (*HealthReportReply, error)
}

type monitorClient struct {
	cc *grpc.ClientConn
}

func NewMonitorClient(cc *grpc.ClientConn) MonitorClient {
	return &monitorClient{cc}
}

func (c *monitorClient) HandleHealthReport(ctx context.Context, in *HealthReport, opts ...grpc.CallOption) (*HealthReportReply, error) {
	out := new(HealthReportReply)
	err := c.cc.Invoke(ctx, "/monitor.Monitor/HandleHealthReport", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MonitorServer is the server API for Monitor service.
type MonitorServer interface {
	HandleHealthReport(context.Context, *HealthReport) (*HealthReportReply, error)
}

func RegisterMonitorServer(s *grpc.Server, srv MonitorServer) {
	s.RegisterService(&_Monitor_serviceDesc, srv)
}

func _Monitor_HandleHealthReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthReport)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonitorServer).HandleHealthReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monitor.Monitor/HandleHealthReport",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonitorServer).HandleHealthReport(ctx, req.(*HealthReport))
	}
	return interceptor(ctx, in, info, handler)
}

var _Monitor_serviceDesc = grpc.ServiceDesc{
	ServiceName: "monitor.Monitor",
	HandlerType: (*MonitorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "HandleHealthReport",
			Handler:    _Monitor_HandleHealthReport_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "monitor.proto",
}
