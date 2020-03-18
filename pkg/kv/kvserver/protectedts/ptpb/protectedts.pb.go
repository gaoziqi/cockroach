// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: kv/kvserver/protectedts/ptpb/protectedts.proto

package ptpb

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import roachpb "github.com/cockroachdb/cockroach/pkg/roachpb"
import hlc "github.com/cockroachdb/cockroach/pkg/util/hlc"

import github_com_cockroachdb_cockroach_pkg_util_uuid "github.com/cockroachdb/cockroach/pkg/util/uuid"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// ProtectionMode defines the semantics of a Record.
type ProtectionMode int32

const (
	// PROTECT_AFTER ensures that all data values live at or after the specified
	// timestamp will be protected from GC.
	PROTECT_AFTER ProtectionMode = 0
)

var ProtectionMode_name = map[int32]string{
	0: "PROTECT_AFTER",
}
var ProtectionMode_value = map[string]int32{
	"PROTECT_AFTER": 0,
}

func (x ProtectionMode) String() string {
	return proto.EnumName(ProtectionMode_name, int32(x))
}
func (ProtectionMode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_protectedts_5535080d2f6ed5b9, []int{0}
}

// Metadata is the system metadata. The metadata is stored explicitly and all
// operations which create or release Records increment the version and update
// the metadata fields accordingly.
//
// The version provides a mechanism for cheap caching and forms the basis of
// the implementation of the Tracker. The Tracker needs to provide a recent
// view of the protectedts subsystem for GC to proceed. The protectedts
// state changes rarely. The timestamp of cached state can by updated by
// merely observing that the version has not changed.
type Metadata struct {
	// Version is incremented whenever a Record is created or removed.
	Version uint64 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	// NumRecords is the number of records which exist in the subsystem.
	NumRecords uint64 `protobuf:"varint,2,opt,name=num_records,json=numRecords,proto3" json:"num_records,omitempty"`
	// NumSpans is the number of spans currently being protected by the
	// protectedts subsystem.
	NumSpans uint64 `protobuf:"varint,3,opt,name=num_spans,json=numSpans,proto3" json:"num_spans,omitempty"`
	// TotalBytes is the number of bytes currently in use by records
	// to store their spans and metadata.
	TotalBytes uint64 `protobuf:"varint,4,opt,name=total_bytes,json=totalBytes,proto3" json:"total_bytes,omitempty"`
}

func (m *Metadata) Reset()         { *m = Metadata{} }
func (m *Metadata) String() string { return proto.CompactTextString(m) }
func (*Metadata) ProtoMessage()    {}
func (*Metadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_protectedts_5535080d2f6ed5b9, []int{0}
}
func (m *Metadata) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Metadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *Metadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Metadata.Merge(dst, src)
}
func (m *Metadata) XXX_Size() int {
	return m.Size()
}
func (m *Metadata) XXX_DiscardUnknown() {
	xxx_messageInfo_Metadata.DiscardUnknown(m)
}

var xxx_messageInfo_Metadata proto.InternalMessageInfo

// Record corresponds to a protected timestamp.
type Record struct {
	// ID uniquely identifies this row.
	ID github_com_cockroachdb_cockroach_pkg_util_uuid.UUID `protobuf:"bytes,1,opt,name=id,proto3,customtype=github.com/cockroachdb/cockroach/pkg/util/uuid.UUID" json:"id"`
	// Timestamp is the timestamp which is protected.
	Timestamp hlc.Timestamp `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp"`
	// Mode specifies whether this record protects all values live at timestamp
	// or all values live at or after that timestamp.
	Mode ProtectionMode `protobuf:"varint,3,opt,name=mode,proto3,enum=cockroach.protectedts.ProtectionMode" json:"mode,omitempty"`
	// MetaType is used to interpret the data stored in Meta.
	// Users of Meta should set a unique value for MetaType which provides enough
	// information to interpret the data in Meta. See the comment on Meta for how
	// these two fields should be used in tandem.
	MetaType string `protobuf:"bytes,4,opt,name=meta_type,json=metaType,proto3" json:"meta_type,omitempty"`
	// Meta is client-provided metadata about the record.
	// This data allows the Record to be correlated with data from another
	// subsystem. For example, this field may contain the ID of a job which
	// created this record. The metadata allows an out-of-band reconciliation
	// process to discover and remove records which no longer correspond to
	// running jobs. Such a mechanism acts as a failsafe against unreliable
	// jobs infrastructure.
	Meta []byte `protobuf:"bytes,5,opt,name=meta,proto3" json:"meta,omitempty"`
	// Verified marks that this Record is known to have successfully provided
	// protection. It is updated after Verification. Updates to this field do not
	// change the Version of the subsystem.
	Verified bool `protobuf:"varint,6,opt,name=verified,proto3" json:"verified,omitempty"`
	// Spans are the spans which this Record protects.
	Spans []roachpb.Span `protobuf:"bytes,7,rep,name=spans,proto3" json:"spans"`
}

func (m *Record) Reset()         { *m = Record{} }
func (m *Record) String() string { return proto.CompactTextString(m) }
func (*Record) ProtoMessage()    {}
func (*Record) Descriptor() ([]byte, []int) {
	return fileDescriptor_protectedts_5535080d2f6ed5b9, []int{1}
}
func (m *Record) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Record) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *Record) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Record.Merge(dst, src)
}
func (m *Record) XXX_Size() int {
	return m.Size()
}
func (m *Record) XXX_DiscardUnknown() {
	xxx_messageInfo_Record.DiscardUnknown(m)
}

var xxx_messageInfo_Record proto.InternalMessageInfo

// State is the complete system state.
type State struct {
	Metadata `protobuf:"bytes,1,opt,name=metadata,proto3,embedded=metadata" json:"metadata"`
	Records  []Record `protobuf:"bytes,2,rep,name=records,proto3" json:"records"`
}

func (m *State) Reset()         { *m = State{} }
func (m *State) String() string { return proto.CompactTextString(m) }
func (*State) ProtoMessage()    {}
func (*State) Descriptor() ([]byte, []int) {
	return fileDescriptor_protectedts_5535080d2f6ed5b9, []int{2}
}
func (m *State) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *State) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *State) XXX_Merge(src proto.Message) {
	xxx_messageInfo_State.Merge(dst, src)
}
func (m *State) XXX_Size() int {
	return m.Size()
}
func (m *State) XXX_DiscardUnknown() {
	xxx_messageInfo_State.DiscardUnknown(m)
}

var xxx_messageInfo_State proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Metadata)(nil), "cockroach.protectedts.Metadata")
	proto.RegisterType((*Record)(nil), "cockroach.protectedts.Record")
	proto.RegisterType((*State)(nil), "cockroach.protectedts.State")
	proto.RegisterEnum("cockroach.protectedts.ProtectionMode", ProtectionMode_name, ProtectionMode_value)
}
func (m *Metadata) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Metadata) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Version != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintProtectedts(dAtA, i, uint64(m.Version))
	}
	if m.NumRecords != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintProtectedts(dAtA, i, uint64(m.NumRecords))
	}
	if m.NumSpans != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintProtectedts(dAtA, i, uint64(m.NumSpans))
	}
	if m.TotalBytes != 0 {
		dAtA[i] = 0x20
		i++
		i = encodeVarintProtectedts(dAtA, i, uint64(m.TotalBytes))
	}
	return i, nil
}

func (m *Record) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Record) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintProtectedts(dAtA, i, uint64(m.ID.Size()))
	n1, err := m.ID.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	dAtA[i] = 0x12
	i++
	i = encodeVarintProtectedts(dAtA, i, uint64(m.Timestamp.Size()))
	n2, err := m.Timestamp.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n2
	if m.Mode != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintProtectedts(dAtA, i, uint64(m.Mode))
	}
	if len(m.MetaType) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintProtectedts(dAtA, i, uint64(len(m.MetaType)))
		i += copy(dAtA[i:], m.MetaType)
	}
	if len(m.Meta) > 0 {
		dAtA[i] = 0x2a
		i++
		i = encodeVarintProtectedts(dAtA, i, uint64(len(m.Meta)))
		i += copy(dAtA[i:], m.Meta)
	}
	if m.Verified {
		dAtA[i] = 0x30
		i++
		if m.Verified {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	if len(m.Spans) > 0 {
		for _, msg := range m.Spans {
			dAtA[i] = 0x3a
			i++
			i = encodeVarintProtectedts(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *State) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *State) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintProtectedts(dAtA, i, uint64(m.Metadata.Size()))
	n3, err := m.Metadata.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n3
	if len(m.Records) > 0 {
		for _, msg := range m.Records {
			dAtA[i] = 0x12
			i++
			i = encodeVarintProtectedts(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func encodeVarintProtectedts(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Metadata) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Version != 0 {
		n += 1 + sovProtectedts(uint64(m.Version))
	}
	if m.NumRecords != 0 {
		n += 1 + sovProtectedts(uint64(m.NumRecords))
	}
	if m.NumSpans != 0 {
		n += 1 + sovProtectedts(uint64(m.NumSpans))
	}
	if m.TotalBytes != 0 {
		n += 1 + sovProtectedts(uint64(m.TotalBytes))
	}
	return n
}

func (m *Record) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.ID.Size()
	n += 1 + l + sovProtectedts(uint64(l))
	l = m.Timestamp.Size()
	n += 1 + l + sovProtectedts(uint64(l))
	if m.Mode != 0 {
		n += 1 + sovProtectedts(uint64(m.Mode))
	}
	l = len(m.MetaType)
	if l > 0 {
		n += 1 + l + sovProtectedts(uint64(l))
	}
	l = len(m.Meta)
	if l > 0 {
		n += 1 + l + sovProtectedts(uint64(l))
	}
	if m.Verified {
		n += 2
	}
	if len(m.Spans) > 0 {
		for _, e := range m.Spans {
			l = e.Size()
			n += 1 + l + sovProtectedts(uint64(l))
		}
	}
	return n
}

func (m *State) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Metadata.Size()
	n += 1 + l + sovProtectedts(uint64(l))
	if len(m.Records) > 0 {
		for _, e := range m.Records {
			l = e.Size()
			n += 1 + l + sovProtectedts(uint64(l))
		}
	}
	return n
}

func sovProtectedts(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozProtectedts(x uint64) (n int) {
	return sovProtectedts(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Metadata) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtectedts
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Metadata: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Metadata: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Version", wireType)
			}
			m.Version = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtectedts
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Version |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NumRecords", wireType)
			}
			m.NumRecords = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtectedts
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NumRecords |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NumSpans", wireType)
			}
			m.NumSpans = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtectedts
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NumSpans |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalBytes", wireType)
			}
			m.TotalBytes = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtectedts
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TotalBytes |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipProtectedts(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProtectedts
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Record) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtectedts
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Record: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Record: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtectedts
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthProtectedts
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ID.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timestamp", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtectedts
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthProtectedts
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Timestamp.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Mode", wireType)
			}
			m.Mode = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtectedts
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Mode |= (ProtectionMode(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MetaType", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtectedts
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProtectedts
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MetaType = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Meta", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtectedts
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthProtectedts
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Meta = append(m.Meta[:0], dAtA[iNdEx:postIndex]...)
			if m.Meta == nil {
				m.Meta = []byte{}
			}
			iNdEx = postIndex
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Verified", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtectedts
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Verified = bool(v != 0)
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Spans", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtectedts
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthProtectedts
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Spans = append(m.Spans, roachpb.Span{})
			if err := m.Spans[len(m.Spans)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProtectedts(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProtectedts
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *State) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtectedts
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: State: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: State: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Metadata", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtectedts
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthProtectedts
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Metadata.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Records", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtectedts
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthProtectedts
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Records = append(m.Records, Record{})
			if err := m.Records[len(m.Records)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProtectedts(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProtectedts
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipProtectedts(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowProtectedts
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowProtectedts
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowProtectedts
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthProtectedts
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowProtectedts
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipProtectedts(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthProtectedts = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowProtectedts   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("kv/kvserver/protectedts/ptpb/protectedts.proto", fileDescriptor_protectedts_5535080d2f6ed5b9)
}

var fileDescriptor_protectedts_5535080d2f6ed5b9 = []byte{
	// 555 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x53, 0xcf, 0x8f, 0xd2, 0x40,
	0x14, 0x6e, 0xa1, 0x40, 0x19, 0x74, 0xa3, 0x13, 0x8d, 0x0d, 0x6a, 0x4b, 0x48, 0x34, 0xe8, 0xa1,
	0x93, 0xc0, 0xc9, 0x83, 0x07, 0x70, 0x31, 0xd9, 0xc3, 0xc6, 0x75, 0x96, 0xbd, 0x78, 0x21, 0xd3,
	0x76, 0x2c, 0x13, 0x68, 0xa7, 0x69, 0xa7, 0x24, 0x9c, 0xbd, 0x78, 0x31, 0xf1, 0x7f, 0xf0, 0x9f,
	0xe1, 0xc8, 0x71, 0xa3, 0x09, 0x51, 0xf8, 0x47, 0xcc, 0x4c, 0xf9, 0xa5, 0xd1, 0xdb, 0x9b, 0xf7,
	0xbe, 0xef, 0xcd, 0x7c, 0xdf, 0x7b, 0x03, 0xdc, 0xe9, 0x1c, 0x4d, 0xe7, 0x19, 0x4d, 0xe7, 0x34,
	0x45, 0x49, 0xca, 0x05, 0xf5, 0x05, 0x0d, 0x44, 0x86, 0x12, 0x91, 0x78, 0xa7, 0x09, 0x57, 0xc6,
	0x1c, 0x3e, 0xf4, 0xb9, 0x3f, 0x4d, 0x39, 0xf1, 0x27, 0xee, 0x49, 0xb1, 0xf9, 0x20, 0xe4, 0x21,
	0x57, 0x08, 0x24, 0xa3, 0x02, 0xdc, 0x7c, 0x12, 0x72, 0x1e, 0xce, 0x28, 0x22, 0x09, 0x43, 0x24,
	0x8e, 0xb9, 0x20, 0x82, 0xf1, 0x78, 0xd7, 0xaa, 0x09, 0x55, 0x9b, 0xc4, 0x43, 0x01, 0x11, 0x64,
	0x97, 0xb3, 0x72, 0xc1, 0x66, 0x68, 0x32, 0xf3, 0x91, 0x60, 0x11, 0xcd, 0x04, 0x89, 0x92, 0xa2,
	0xd2, 0xfe, 0xa4, 0x03, 0xf3, 0x92, 0x0a, 0x22, 0xc1, 0xd0, 0x02, 0xb5, 0x39, 0x4d, 0x33, 0xc6,
	0x63, 0x4b, 0x6f, 0xe9, 0x1d, 0x03, 0xef, 0x8f, 0xd0, 0x01, 0x8d, 0x38, 0x8f, 0xc6, 0x29, 0xf5,
	0x79, 0x1a, 0x64, 0x56, 0x49, 0x55, 0x41, 0x9c, 0x47, 0xb8, 0xc8, 0xc0, 0xc7, 0xa0, 0x2e, 0x01,
	0x59, 0x42, 0xe2, 0xcc, 0x2a, 0xab, 0xb2, 0x19, 0xe7, 0xd1, 0xb5, 0x3c, 0x4b, 0xb6, 0xe0, 0x82,
	0xcc, 0xc6, 0xde, 0x42, 0xd0, 0xcc, 0x32, 0x0a, 0xb6, 0x4a, 0x0d, 0x64, 0xa6, 0xfd, 0xa3, 0x04,
	0xaa, 0x45, 0x27, 0xf8, 0x1e, 0x94, 0x58, 0xa0, 0xae, 0xbf, 0x33, 0xe8, 0x2f, 0xd7, 0x8e, 0xf6,
	0x7d, 0xed, 0xf4, 0x42, 0x26, 0x26, 0xb9, 0xe7, 0xfa, 0x3c, 0x42, 0x07, 0xa3, 0x02, 0xef, 0x18,
	0xa3, 0x64, 0x1a, 0x22, 0xa5, 0x31, 0xcf, 0x59, 0xe0, 0xde, 0xdc, 0x5c, 0x9c, 0x6f, 0xd6, 0x4e,
	0xe9, 0xe2, 0x1c, 0x97, 0x58, 0x00, 0xfb, 0xa0, 0x7e, 0x90, 0xad, 0x9e, 0xde, 0xe8, 0x3e, 0x75,
	0x8f, 0x86, 0x4b, 0x9e, 0x3b, 0x99, 0xf9, 0xee, 0x68, 0x0f, 0x1a, 0x18, 0xf2, 0x62, 0x7c, 0x64,
	0xc1, 0x57, 0xc0, 0x88, 0x78, 0x40, 0x95, 0xb2, 0xb3, 0xee, 0x33, 0xf7, 0x9f, 0xe3, 0x72, 0xaf,
	0x8a, 0x98, 0xf1, 0xf8, 0x92, 0x07, 0x14, 0x2b, 0x8a, 0x74, 0x26, 0xa2, 0x82, 0x8c, 0xc5, 0x22,
	0xa1, 0x4a, 0x7a, 0x1d, 0x9b, 0x32, 0x31, 0x5a, 0x24, 0x14, 0x42, 0x60, 0xc8, 0xd8, 0xaa, 0x48,
	0xbd, 0x58, 0xc5, 0xb0, 0x09, 0xcc, 0x39, 0x4d, 0xd9, 0x47, 0x46, 0x03, 0xab, 0xda, 0xd2, 0x3b,
	0x26, 0x3e, 0x9c, 0x61, 0x0f, 0x54, 0x0a, 0x8b, 0x6b, 0xad, 0x72, 0xa7, 0xd1, 0x7d, 0x74, 0xf2,
	0x90, 0xdd, 0xd8, 0x5d, 0x69, 0xf9, 0x4e, 0x40, 0x81, 0x6d, 0x7f, 0xd1, 0x41, 0xe5, 0x5a, 0x10,
	0x41, 0xe1, 0x10, 0xa8, 0xab, 0xe5, 0xb0, 0x95, 0xc5, 0x8d, 0xae, 0xf3, 0x1f, 0x29, 0xfb, 0x9d,
	0x18, 0x98, 0xb2, 0xd3, 0x6a, 0xed, 0xe8, 0xf8, 0x40, 0x85, 0xaf, 0x41, 0xed, 0xb8, 0x09, 0xe5,
	0xbf, 0xec, 0x3c, 0xed, 0x52, 0xcc, 0x74, 0xf7, 0x9a, 0x3d, 0xe7, 0xe5, 0x0b, 0x70, 0xf6, 0xa7,
	0x53, 0xf0, 0x3e, 0xb8, 0x7b, 0x85, 0xdf, 0x8d, 0x86, 0x6f, 0x46, 0xe3, 0xfe, 0xdb, 0xd1, 0x10,
	0xdf, 0xd3, 0x9a, 0xc6, 0xe7, 0x6f, 0xb6, 0x36, 0x78, 0xbe, 0xfc, 0x65, 0x6b, 0xcb, 0x8d, 0xad,
	0xaf, 0x36, 0xb6, 0x7e, 0xbb, 0xb1, 0xf5, 0x9f, 0x1b, 0x5b, 0xff, 0xba, 0xb5, 0xb5, 0xd5, 0xd6,
	0xd6, 0x6e, 0xb7, 0xb6, 0xf6, 0xc1, 0x90, 0x3f, 0xca, 0xab, 0xaa, 0x6d, 0xee, 0xfd, 0x0e, 0x00,
	0x00, 0xff, 0xff, 0xf9, 0x47, 0xda, 0x7a, 0x78, 0x03, 0x00, 0x00,
}
