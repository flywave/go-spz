package spz

import "encoding/binary"

const (
	HeaderSizeSpz = 16
	SPZ_MAGIC     = 0x5053474e // NGSP = Niantic gaussian splat
	CMask         = 0x1FF      // 9 bits mask
	SQRT1_2       = 0.7071067811865476
)

// SpzData represents the header and data of an SPZ file
type SpzData struct {
	// Header fields
	Magic          uint32
	Version        uint32
	NumPoints      uint32
	ShDegree       uint8
	FractionalBits uint8
	Flags          uint8
	Reserved       uint8

	// Data fields
	Data []*SplatData
}

// ToBytes converts SpzData header to bytes
func (h *SpzData) ToBytes() []byte {
	bts := make([]byte, HeaderSizeSpz)
	binary.LittleEndian.PutUint32(bts[0:4], h.Magic)
	binary.LittleEndian.PutUint32(bts[4:8], h.Version)
	binary.LittleEndian.PutUint32(bts[8:12], h.NumPoints)
	bts[12] = h.ShDegree
	bts[13] = h.FractionalBits
	bts[14] = h.Flags
	bts[15] = h.Reserved
	return bts
}

// SplatData represents a single splat data point
type SplatData struct {
	PositionX float32
	PositionY float32
	PositionZ float32
	ScaleX    float32
	ScaleY    float32
	ScaleZ    float32
	RotationW uint8
	RotationX uint8
	RotationY uint8
	RotationZ uint8
	ColorR    uint8
	ColorG    uint8
	ColorB    uint8
	ColorA    uint8
	SH1       []byte
	SH2       []byte
	SH3       []byte
}

// ParseSpzHeader parses the header of an SPZ file
func ParseSpzHeader(data []byte) (*SpzData, error) {
	if len(data) < HeaderSizeSpz {
		return nil, &SpzError{"Invalid SPZ file: insufficient data for header"}
	}

	spzData := &SpzData{}
	spzData.Magic = uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	spzData.Version = uint32(data[4]) | uint32(data[5])<<8 | uint32(data[6])<<16 | uint32(data[7])<<24
	spzData.NumPoints = uint32(data[8]) | uint32(data[9])<<8 | uint32(data[10])<<16 | uint32(data[11])<<24
	spzData.ShDegree = data[12]
	spzData.FractionalBits = data[13]
	spzData.Flags = data[14]
	spzData.Reserved = data[15]

	// Validate header
	if spzData.Magic != SPZ_MAGIC {
		return nil, &SpzError{"Invalid SPZ file: magic number mismatch"}
	}
	if spzData.Version < 2 || spzData.Version > 3 {
		return nil, &SpzError{"Unsupported SPZ version: " + string(rune(spzData.Version))}
	}
	if spzData.ShDegree > 3 {
		return nil, &SpzError{"Unsupported SH degree: " + string(rune(spzData.ShDegree))}
	}
	if spzData.FractionalBits != 12 {
		return nil, &SpzError{"Unsupported fractional bits: " + string(rune(spzData.FractionalBits))}
	}

	return spzData, nil
}
