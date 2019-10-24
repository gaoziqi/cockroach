// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sql/stats/histogram.proto

package stats

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import github_com_cockroachdb_cockroach_pkg_sql_types "github.com/cockroachdb/cockroach/pkg/sql/types"

import encoding_binary "encoding/binary"

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

// HistogramData encodes the data for a histogram, which captures the
// distribution of values on a specific column.
type HistogramData struct {
	// Value type for the column.
	ColumnType github_com_cockroachdb_cockroach_pkg_sql_types.T `protobuf:"bytes,2,opt,name=column_type,json=columnType,proto3,customtype=github.com/cockroachdb/cockroach/pkg/sql/types.T" json:"column_type"`
	// Histogram buckets. Note that NULL values are excluded from the
	// histogram.
	Buckets []HistogramData_Bucket `protobuf:"bytes,1,rep,name=buckets,proto3" json:"buckets"`
}

func (m *HistogramData) Reset()         { *m = HistogramData{} }
func (m *HistogramData) String() string { return proto.CompactTextString(m) }
func (*HistogramData) ProtoMessage()    {}
func (*HistogramData) Descriptor() ([]byte, []int) {
	return fileDescriptor_histogram_37d853d20c5166cc, []int{0}
}
func (m *HistogramData) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *HistogramData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *HistogramData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HistogramData.Merge(dst, src)
}
func (m *HistogramData) XXX_Size() int {
	return m.Size()
}
func (m *HistogramData) XXX_DiscardUnknown() {
	xxx_messageInfo_HistogramData.DiscardUnknown(m)
}

var xxx_messageInfo_HistogramData proto.InternalMessageInfo

type HistogramData_Bucket struct {
	// The estimated number of values that are equal to upper_bound.
	NumEq int64 `protobuf:"varint,1,opt,name=num_eq,json=numEq,proto3" json:"num_eq,omitempty"`
	// The estimated number of values in the bucket (excluding those
	// that are equal to upper_bound). Splitting the count into two
	// makes the histogram effectively equivalent to a histogram with
	// twice as many buckets, with every other bucket containing a
	// single value. This might be particularly advantageous if the
	// histogram algorithm makes sure the top "heavy hitters" (most
	// frequent elements) are bucket boundaries (similar to a
	// compressed histogram).
	NumRange int64 `protobuf:"varint,2,opt,name=num_range,json=numRange,proto3" json:"num_range,omitempty"`
	// The estimated number of distinct values in the bucket (excluding
	// those that are equal to upper_bound). This is a floating point
	// value because it is estimated by distributing the known distinct
	// count for the column among the buckets, in proportion to the number
	// of rows in each bucket. This value is in fact derived from the rest
	// of the data, but is included to avoid re-computing it later.
	DistinctRange float64 `protobuf:"fixed64,4,opt,name=distinct_range,json=distinctRange,proto3" json:"distinct_range,omitempty"`
	// The upper boundary of the bucket. The column values for the upper bound
	// are encoded using the ascending key encoding of the column type.
	UpperBound []byte `protobuf:"bytes,3,opt,name=upper_bound,json=upperBound,proto3" json:"upper_bound,omitempty"`
}

func (m *HistogramData_Bucket) Reset()         { *m = HistogramData_Bucket{} }
func (m *HistogramData_Bucket) String() string { return proto.CompactTextString(m) }
func (*HistogramData_Bucket) ProtoMessage()    {}
func (*HistogramData_Bucket) Descriptor() ([]byte, []int) {
	return fileDescriptor_histogram_37d853d20c5166cc, []int{0, 0}
}
func (m *HistogramData_Bucket) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *HistogramData_Bucket) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *HistogramData_Bucket) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HistogramData_Bucket.Merge(dst, src)
}
func (m *HistogramData_Bucket) XXX_Size() int {
	return m.Size()
}
func (m *HistogramData_Bucket) XXX_DiscardUnknown() {
	xxx_messageInfo_HistogramData_Bucket.DiscardUnknown(m)
}

var xxx_messageInfo_HistogramData_Bucket proto.InternalMessageInfo

func init() {
	proto.RegisterType((*HistogramData)(nil), "cockroach.sql.stats.HistogramData")
	proto.RegisterType((*HistogramData_Bucket)(nil), "cockroach.sql.stats.HistogramData.Bucket")
}
func (m *HistogramData) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *HistogramData) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Buckets) > 0 {
		for _, msg := range m.Buckets {
			dAtA[i] = 0xa
			i++
			i = encodeVarintHistogram(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	dAtA[i] = 0x12
	i++
	i = encodeVarintHistogram(dAtA, i, uint64(m.ColumnType.Size()))
	n1, err := m.ColumnType.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	return i, nil
}

func (m *HistogramData_Bucket) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *HistogramData_Bucket) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.NumEq != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintHistogram(dAtA, i, uint64(m.NumEq))
	}
	if m.NumRange != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintHistogram(dAtA, i, uint64(m.NumRange))
	}
	if len(m.UpperBound) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintHistogram(dAtA, i, uint64(len(m.UpperBound)))
		i += copy(dAtA[i:], m.UpperBound)
	}
	if m.DistinctRange != 0 {
		dAtA[i] = 0x21
		i++
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.DistinctRange))))
		i += 8
	}
	return i, nil
}

func encodeVarintHistogram(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *HistogramData) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Buckets) > 0 {
		for _, e := range m.Buckets {
			l = e.Size()
			n += 1 + l + sovHistogram(uint64(l))
		}
	}
	l = m.ColumnType.Size()
	n += 1 + l + sovHistogram(uint64(l))
	return n
}

func (m *HistogramData_Bucket) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.NumEq != 0 {
		n += 1 + sovHistogram(uint64(m.NumEq))
	}
	if m.NumRange != 0 {
		n += 1 + sovHistogram(uint64(m.NumRange))
	}
	l = len(m.UpperBound)
	if l > 0 {
		n += 1 + l + sovHistogram(uint64(l))
	}
	if m.DistinctRange != 0 {
		n += 9
	}
	return n
}

func sovHistogram(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozHistogram(x uint64) (n int) {
	return sovHistogram(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *HistogramData) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowHistogram
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
			return fmt.Errorf("proto: HistogramData: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: HistogramData: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Buckets", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHistogram
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
				return ErrInvalidLengthHistogram
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Buckets = append(m.Buckets, HistogramData_Bucket{})
			if err := m.Buckets[len(m.Buckets)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ColumnType", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHistogram
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
				return ErrInvalidLengthHistogram
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ColumnType.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipHistogram(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthHistogram
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
func (m *HistogramData_Bucket) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowHistogram
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
			return fmt.Errorf("proto: Bucket: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Bucket: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NumEq", wireType)
			}
			m.NumEq = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHistogram
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NumEq |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NumRange", wireType)
			}
			m.NumRange = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHistogram
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NumRange |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field UpperBound", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHistogram
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
				return ErrInvalidLengthHistogram
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.UpperBound = append(m.UpperBound[:0], dAtA[iNdEx:postIndex]...)
			if m.UpperBound == nil {
				m.UpperBound = []byte{}
			}
			iNdEx = postIndex
		case 4:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field DistinctRange", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.DistinctRange = float64(math.Float64frombits(v))
		default:
			iNdEx = preIndex
			skippy, err := skipHistogram(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthHistogram
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
func skipHistogram(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowHistogram
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
					return 0, ErrIntOverflowHistogram
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
					return 0, ErrIntOverflowHistogram
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
				return 0, ErrInvalidLengthHistogram
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowHistogram
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
				next, err := skipHistogram(dAtA[start:])
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
	ErrInvalidLengthHistogram = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowHistogram   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("sql/stats/histogram.proto", fileDescriptor_histogram_37d853d20c5166cc)
}

var fileDescriptor_histogram_37d853d20c5166cc = []byte{
	// 339 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0xb1, 0x4e, 0xeb, 0x30,
	0x18, 0x85, 0xe3, 0xa6, 0xed, 0xbd, 0xd7, 0xbd, 0x65, 0x08, 0x20, 0x85, 0x22, 0xb9, 0x11, 0x12,
	0x22, 0x2c, 0x0e, 0x82, 0x85, 0x39, 0x02, 0x09, 0x56, 0xab, 0x0b, 0x2c, 0x55, 0xe2, 0x46, 0x69,
	0xd4, 0xc6, 0x4e, 0x62, 0x7b, 0xe8, 0xce, 0x03, 0xf0, 0x40, 0x3c, 0x40, 0xc7, 0x8e, 0x15, 0x43,
	0x05, 0xe9, 0x8b, 0x20, 0x3b, 0x14, 0x84, 0xc4, 0xf6, 0xfb, 0xf8, 0x7c, 0xe7, 0xd8, 0x3f, 0x3c,
	0x12, 0xe5, 0x3c, 0x10, 0x32, 0x92, 0x22, 0x98, 0x66, 0x42, 0xf2, 0xb4, 0x8a, 0x72, 0x5c, 0x54,
	0x5c, 0x72, 0x67, 0x9f, 0x72, 0x3a, 0xab, 0x78, 0x44, 0xa7, 0x58, 0x94, 0x73, 0x6c, 0x4c, 0x83,
	0x83, 0x94, 0xa7, 0xdc, 0xdc, 0x07, 0x7a, 0x6a, 0xac, 0x27, 0x2f, 0x2d, 0xd8, 0xbf, 0xdb, 0xe1,
	0x37, 0x91, 0x8c, 0x9c, 0x7b, 0xf8, 0x27, 0x56, 0x74, 0x96, 0x48, 0xe1, 0x02, 0xcf, 0xf6, 0x7b,
	0x97, 0xe7, 0xf8, 0x97, 0x38, 0xfc, 0x03, 0xc2, 0xa1, 0x21, 0xc2, 0xf6, 0x72, 0x33, 0xb4, 0xc8,
	0x8e, 0x77, 0x1e, 0x60, 0x8f, 0xf2, 0xb9, 0xca, 0xd9, 0x58, 0x2e, 0x8a, 0xc4, 0x6d, 0x79, 0xc0,
	0xff, 0x1f, 0x5e, 0x6b, 0xcf, 0xeb, 0x66, 0x78, 0x91, 0x66, 0x72, 0xaa, 0x62, 0x4c, 0x79, 0x1e,
	0x7c, 0x15, 0x4c, 0xe2, 0xef, 0x39, 0x28, 0x66, 0x69, 0xa0, 0x3f, 0xa9, 0x61, 0x81, 0x47, 0x04,
	0x36, 0x61, 0xa3, 0x45, 0x91, 0x0c, 0x9e, 0x00, 0xec, 0x36, 0xa5, 0xce, 0x21, 0xec, 0x32, 0x95,
	0x8f, 0x93, 0xd2, 0x05, 0x1e, 0xf0, 0x6d, 0xd2, 0x61, 0x2a, 0xbf, 0x2d, 0x9d, 0x63, 0xf8, 0x4f,
	0xcb, 0x55, 0xc4, 0xd2, 0xa6, 0xda, 0x26, 0x7f, 0x99, 0xca, 0x89, 0x3e, 0x3b, 0x43, 0xd8, 0x53,
	0x45, 0x91, 0x54, 0xe3, 0x98, 0x2b, 0x36, 0x71, 0x6d, 0xfd, 0x32, 0x02, 0x8d, 0x14, 0x6a, 0xc5,
	0x39, 0x85, 0x7b, 0x93, 0x4c, 0xc8, 0x8c, 0x51, 0xf9, 0x19, 0xd1, 0xf6, 0x80, 0x0f, 0x48, 0x7f,
	0xa7, 0x9a, 0x9c, 0xf0, 0x6c, 0xf9, 0x8e, 0xac, 0x65, 0x8d, 0xc0, 0xaa, 0x46, 0x60, 0x5d, 0x23,
	0xf0, 0x56, 0x23, 0xf0, 0xbc, 0x45, 0xd6, 0x6a, 0x8b, 0xac, 0xf5, 0x16, 0x59, 0x8f, 0x1d, 0xb3,
	0xae, 0xb8, 0x6b, 0xd6, 0x7d, 0xf5, 0x11, 0x00, 0x00, 0xff, 0xff, 0x0a, 0x3f, 0x18, 0xc2, 0xb6,
	0x01, 0x00, 0x00,
}
