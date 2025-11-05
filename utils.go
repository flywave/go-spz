package spz

import (
	"encoding/binary"
	"math"
)

const (
	SH_C0       = 0.28209479177387814
	COLOR_SCALE = 2.0
)

func clipFloat32(v float64) float32 {
	if v < -math.MaxFloat32 {
		return -math.MaxFloat32
	}
	if v > math.MaxFloat32 {
		return math.MaxFloat32
	}
	return float32(v)
}

func clipUint8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func spzDecodeScale(val uint8) float32 {
	return float32(val)/16.0 - 10.0
}

func spzDecodePosition(bts []byte, fractionalBits uint8) float32 {
	scale := 1.0 / float64(int(1<<fractionalBits))
	fixed32 := int32(bts[0]) | int32(bts[1])<<8 | int32(bts[2])<<16
	if fixed32&0x800000 != 0 {
		// fixed32 |= int32(-1) << 24 // 如果符号位为1，将高8位填充为1
		fixed32 |= -0x1000000
	}
	return clipFloat32(float64(fixed32) * scale)
}

func spzDecodeRotations(rx uint8, ry uint8, rz uint8) (uint8, uint8, uint8, uint8) {
	r1 := float64(rx)/127.5 - 1.0
	r2 := float64(ry)/127.5 - 1.0
	r3 := float64(rz)/127.5 - 1.0
	r0 := math.Sqrt(math.Max(0.0, 1.0-(r1*r1+r2*r2+r3*r3)))
	return clipUint8(r0*128.0 + 128.0), clipUint8(r1*128.0 + 128.0), clipUint8(r2*128.0 + 128.0), clipUint8(r3*128.0 + 128.0)
}

func spzDecodeRotationsV3(bs []byte) (uint8, uint8, uint8, uint8) {
	comp := binary.LittleEndian.Uint32(bs)
	index := int(comp >> 30)
	remaining := comp
	sumSquares := 0.0
	rotation := []float64{0.0, 0.0, 0.0, 0.0}

	for i := 3; i >= 0; i-- {
		if i != index {
			magnitude := float64(remaining & CMask)
			negbit := (remaining >> 9) & 0x1
			remaining = remaining >> 10

			rotation[i] = SQRT1_2 * (magnitude / float64(CMask))
			if negbit == 1 {
				rotation[i] = -rotation[i]
			}

			sumSquares += rotation[i] * rotation[i]
		}
	}

	rotation[index] = math.Sqrt(math.Max(1.0-sumSquares, 0))

	r0, r1, r2, r3 := rotation[0], rotation[1], rotation[2], rotation[3]
	return clipUint8(r0*128.0 + 128.0), clipUint8(r1*128.0 + 128.0), clipUint8(r2*128.0 + 128.0), clipUint8(r3*128.0 + 128.0)
}

// spzDecodeColor decodes color value (inverse of spzEncodeColor)
func spzDecodeColor(val uint8) uint8 {
	const SH_C0 = 0.28209479177387814
	const COLOR_SCALE = 2.0
	// Decode from spz format
	fColor := (float64(val)/255.0 - 0.5) / COLOR_SCALE
	// Restore original color
	original := (fColor*SH_C0 + 0.5) * 255.0
	return clipUint8(original)
}

// spzDecodeSH1 decodes SH1 values (inverse of spzEncodeSH1)
func spzDecodeSH1(val uint8) uint8 {
	// The encoding quantizes to multiples of 8, just return the value
	return val
}

// spzDecodeSH23 decodes SH2 and SH3 values (inverse of spzEncodeSH23)
func spzDecodeSH23(val uint8) uint8 {
	// The encoding quantizes to multiples of 16, just return the value
	return val
}

// clipUint8Round clips and rounds a float64 value to uint8 range
func clipUint8Round(x float64) uint8 {
	return uint8(math.Max(0, math.Min(255, math.Round(x))))
}

// uint32ToBytes converts uint32 to 4 bytes in little endian
func uint32ToBytes(val uint32) []byte {
	bts := make([]byte, 4)
	binary.LittleEndian.PutUint32(bts, val)
	return bts
}

// encodeFloat32ToBytes3 converts float32 to 3 bytes (24-bit fixed point)
func encodeFloat32ToBytes3(f float32) []byte {
	fixed32 := int32(math.Round(float64(f) * 4096))

	return []byte{
		byte(fixed32 & 0xFF),
		byte((fixed32 >> 8) & 0xFF),
		byte((fixed32 >> 16) & 0xFF),
	}
}

// spzEncodePosition encodes position value to 3 bytes
func spzEncodePosition(val float32) []byte {
	return encodeFloat32ToBytes3(val)
}

// spzEncodeColor encodes color value
func spzEncodeColor(val uint8) uint8 {
	fColor := (float64(val)/255.0 - 0.5) / SH_C0
	return clipUint8Round(fColor*(COLOR_SCALE*255.0) + (0.5 * 255.0))
}

// spzEncodeScale encodes scale value
func spzEncodeScale(val float32) uint8 {
	return clipUint8Round((float64(val) + 10.0) * 16.0)
}

// spzEncodeRotations encodes rotation for version 2
func spzEncodeRotations(rw uint8, rx uint8, ry uint8, rz uint8) []byte {
	r0 := float64(rw)/128.0 - 1.0
	r1 := float64(rx)/128.0 - 1.0
	r2 := float64(ry)/128.0 - 1.0
	r3 := float64(rz)/128.0 - 1.0
	if r0 < 0 {
		r0, r1, r2, r3 = -r0, -r1, -r2, -r3
	}
	qlen := math.Sqrt(r0*r0 + r1*r1 + r2*r2 + r3*r3)
	return []byte{
		clipUint8Round((r1/qlen)*127.5 + 127.5),
		clipUint8Round((r2/qlen)*127.5 + 127.5),
		clipUint8Round((r3/qlen)*127.5 + 127.5),
	}
}

// spzEncodeRotationsV3 encodes rotation for version 3
func spzEncodeRotationsV3(rw uint8, rx uint8, ry uint8, rz uint8) []byte {
	r0 := float64(rw)/128.0 - 1.0
	r1 := float64(rx)/128.0 - 1.0
	r2 := float64(ry)/128.0 - 1.0
	r3 := float64(rz)/128.0 - 1.0
	qlen := math.Sqrt(r0*r0 + r1*r1 + r2*r2 + r3*r3)
	rotation := []float64{r0 / qlen, r1 / qlen, r2 / qlen, r3 / qlen}

	index := 0
	for i := 1; i < 4; i++ {
		if math.Abs(rotation[index]) < math.Abs(rotation[i]) {
			index = i
		}
	}
	if rotation[index] < 0 {
		rotation[0], rotation[1], rotation[2], rotation[3] = -rotation[0], -rotation[1], -rotation[2], -rotation[3]
	}

	remaining := uint32(index)
	for i := 0; i < 4; i++ {
		if i != index {
			signBit := uint32(0)
			if rotation[i] < 0 {
				signBit = 1
			}

			component := math.Abs(rotation[i]) / SQRT1_2
			magnitude := uint32(int64(float64(CMask)*component + 0.5))
			remaining = (remaining << 10) | (signBit << 9) | magnitude
		}
	}

	return uint32ToBytes(remaining)
}

// spzEncodeSH1 encodes SH1 values
func spzEncodeSH1(encodeSHval uint8) uint8 {
	q := math.Floor((float64(encodeSHval)+4.0)/8.0) * 8.0
	return clipUint8(q)
}

// spzEncodeSH23 encodes SH2 and SH3 values
func spzEncodeSH23(encodeSHval uint8) uint8 {
	q := math.Floor((float64(encodeSHval)+8.0)/16.0) * 16.0
	return clipUint8(q)
}

// encodeSplatSH encodes SH value from float64
func encodeSplatSH(val float64) uint8 {
	return clipUint8(math.Round(val*128.0) + 128.0)
}
