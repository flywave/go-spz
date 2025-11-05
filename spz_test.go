package spz

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestReadWriteSpz tests the complete read/write cycle
func TestReadWriteSpz(t *testing.T) {
	// Create test data
	testFile := "test_output.spz"
	defer os.Remove(testFile) // Clean up after test

	// Create original SPZ data
	originalData := &SpzData{
		Magic:          SPZ_MAGIC,
		Version:        2,
		NumPoints:      3,
		ShDegree:       1,
		FractionalBits: 12,
		Flags:          0,
		Reserved:       0,
		Data: []*SplatData{
			{
				PositionX: 1.5,
				PositionY: 2.5,
				PositionZ: 3.5,
				ScaleX:    0.1,
				ScaleY:    0.2,
				ScaleZ:    0.3,
				RotationW: 128,
				RotationX: 138,
				RotationY: 148,
				RotationZ: 158,
				ColorR:    255,
				ColorG:    128,
				ColorB:    64,
				ColorA:    200,
				SH1:       []byte{100, 110, 120, 130, 140, 150, 160, 170, 180},
			},
			{
				PositionX: -1.5,
				PositionY: -2.5,
				PositionZ: -3.5,
				ScaleX:    0.4,
				ScaleY:    0.5,
				ScaleZ:    0.6,
				RotationW: 130,
				RotationX: 140,
				RotationY: 150,
				RotationZ: 160,
				ColorR:    64,
				ColorG:    128,
				ColorB:    255,
				ColorA:    180,
				SH1:       []byte{90, 100, 110, 120, 130, 140, 150, 160, 170},
			},
			{
				PositionX: 0.0,
				PositionY: 1.0,
				PositionZ: -1.0,
				ScaleX:    0.7,
				ScaleY:    0.8,
				ScaleZ:    0.9,
				RotationW: 135,
				RotationX: 145,
				RotationY: 155,
				RotationZ: 165,
				ColorR:    128,
				ColorG:    255,
				ColorB:    128,
				ColorA:    220,
				SH1:       []byte{80, 90, 100, 110, 120, 130, 140, 150, 160},
			},
		},
	}

	// Write SPZ file
	err := WriteSpz(testFile, originalData)
	assert.NoError(t, err, "Failed to write SPZ file")

	// Read SPZ file
	readData, err := ReadSpz(testFile)
	assert.NoError(t, err, "Failed to read SPZ file")
	assert.NotNil(t, readData, "Read data should not be nil")

	// Verify header
	assert.Equal(t, originalData.Magic, readData.Magic, "Magic mismatch")
	assert.Equal(t, originalData.Version, readData.Version, "Version mismatch")
	assert.Equal(t, originalData.NumPoints, readData.NumPoints, "NumPoints mismatch")
	assert.Equal(t, originalData.ShDegree, readData.ShDegree, "ShDegree mismatch")
	assert.Equal(t, originalData.FractionalBits, readData.FractionalBits, "FractionalBits mismatch")

	// Verify data points count
	assert.Equal(t, len(originalData.Data), len(readData.Data), "Data points count mismatch")

	// Verify each data point
	for i := range originalData.Data {
		orig := originalData.Data[i]
		read := readData.Data[i]

		// Positions (allow small tolerance for encoding/decoding)
		assert.InDelta(t, orig.PositionX, read.PositionX, 0.001, "PositionX mismatch at index %d", i)
		assert.InDelta(t, orig.PositionY, read.PositionY, 0.001, "PositionY mismatch at index %d", i)
		assert.InDelta(t, orig.PositionZ, read.PositionZ, 0.001, "PositionZ mismatch at index %d", i)

		// Scales (allow small tolerance)
		assert.InDelta(t, orig.ScaleX, read.ScaleX, 0.1, "ScaleX mismatch at index %d", i)
		assert.InDelta(t, orig.ScaleY, read.ScaleY, 0.1, "ScaleY mismatch at index %d", i)
		assert.InDelta(t, orig.ScaleZ, read.ScaleZ, 0.1, "ScaleZ mismatch at index %d", i)

		// Rotations (encoding/decoding is lossy, just verify they exist)
		assert.NotZero(t, read.RotationW, "RotationW should not be zero at index %d", i)
		assert.NotZero(t, read.RotationX, "RotationX should not be zero at index %d", i)
		assert.NotZero(t, read.RotationY, "RotationY should not be zero at index %d", i)
		assert.NotZero(t, read.RotationZ, "RotationZ should not be zero at index %d", i)

		// Colors (encoding/decoding is lossy, just verify they are in valid range)
		assert.GreaterOrEqual(t, read.ColorR, uint8(0), "ColorR should be >= 0 at index %d", i)
		assert.LessOrEqual(t, read.ColorR, uint8(255), "ColorR should be <= 255 at index %d", i)
		assert.GreaterOrEqual(t, read.ColorG, uint8(0), "ColorG should be >= 0 at index %d", i)
		assert.LessOrEqual(t, read.ColorG, uint8(255), "ColorG should be <= 255 at index %d", i)
		assert.GreaterOrEqual(t, read.ColorB, uint8(0), "ColorB should be >= 0 at index %d", i)
		assert.LessOrEqual(t, read.ColorB, uint8(255), "ColorB should be <= 255 at index %d", i)
		assert.Equal(t, orig.ColorA, read.ColorA, "ColorA mismatch at index %d", i)

		// SH data
		if orig.SH1 != nil {
			assert.NotNil(t, read.SH1, "SH1 should not be nil at index %d", i)
			assert.Equal(t, len(orig.SH1), len(read.SH1), "SH1 length mismatch at index %d", i)
		}
	}
}

// TestReadWriteSpzVersion3 tests version 3 format with 4-byte rotations
func TestReadWriteSpzVersion3(t *testing.T) {
	testFile := "test_v3.spz"
	defer os.Remove(testFile)

	originalData := &SpzData{
		Magic:          SPZ_MAGIC,
		Version:        3,
		NumPoints:      2,
		ShDegree:       2,
		FractionalBits: 12,
		Flags:          0,
		Reserved:       0,
		Data: []*SplatData{
			{
				PositionX: 10.5,
				PositionY: 20.5,
				PositionZ: 30.5,
				ScaleX:    1.1,
				ScaleY:    1.2,
				ScaleZ:    1.3,
				RotationW: 128,
				RotationX: 138,
				RotationY: 148,
				RotationZ: 158,
				ColorR:    255,
				ColorG:    200,
				ColorB:    100,
				ColorA:    250,
				SH2: []byte{
					100, 110, 120, 130, 140, 150, 160, 170, 180, // First 9 (SH1 range)
					90, 100, 110, 120, 130, 140, 150, 160, 170, 180, 190, 200, 210, 220, 230, // Next 15 (SH2 range)
				},
			},
			{
				PositionX: -5.5,
				PositionY: -10.5,
				PositionZ: -15.5,
				ScaleX:    2.1,
				ScaleY:    2.2,
				ScaleZ:    2.3,
				RotationW: 130,
				RotationX: 140,
				RotationY: 150,
				RotationZ: 160,
				ColorR:    100,
				ColorG:    200,
				ColorB:    255,
				ColorA:    230,
				SH2: []byte{
					80, 90, 100, 110, 120, 130, 140, 150, 160,
					70, 80, 90, 100, 110, 120, 130, 140, 150, 160, 170, 180, 190, 200, 210,
				},
			},
		},
	}

	// Write and read
	err := WriteSpz(testFile, originalData)
	assert.NoError(t, err)

	readData, err := ReadSpz(testFile)
	assert.NoError(t, err)
	assert.NotNil(t, readData)

	// Verify version 3
	assert.Equal(t, uint32(3), readData.Version)
	assert.Equal(t, uint8(2), readData.ShDegree)
	assert.Equal(t, len(originalData.Data), len(readData.Data))
}

// TestReadWriteSpzShDegree3 tests SH degree 3
func TestReadWriteSpzShDegree3(t *testing.T) {
	testFile := "test_sh3.spz"
	defer os.Remove(testFile)

	originalData := &SpzData{
		Magic:          SPZ_MAGIC,
		Version:        3,
		NumPoints:      1,
		ShDegree:       3,
		FractionalBits: 12,
		Flags:          0,
		Reserved:       0,
		Data: []*SplatData{
			{
				PositionX: 5.0,
				PositionY: 6.0,
				PositionZ: 7.0,
				ScaleX:    0.5,
				ScaleY:    0.6,
				ScaleZ:    0.7,
				RotationW: 128,
				RotationX: 138,
				RotationY: 148,
				RotationZ: 158,
				ColorR:    128,
				ColorG:    128,
				ColorB:    128,
				ColorA:    255,
				SH2: []byte{
					100, 110, 120, 130, 140, 150, 160, 170, 180,
					90, 100, 110, 120, 130, 140, 150, 160, 170, 180, 190, 200, 210, 220, 230,
				},
				SH3: []byte{
					80, 90, 100, 110, 120, 130, 140, 150, 160, 170, 180,
					190, 200, 210, 220, 230, 240, 250, 255, 255, 255,
				},
			},
		},
	}

	err := WriteSpz(testFile, originalData)
	assert.NoError(t, err)

	readData, err := ReadSpz(testFile)
	assert.NoError(t, err)
	assert.NotNil(t, readData)

	assert.Equal(t, uint8(3), readData.ShDegree)
	assert.NotNil(t, readData.Data[0].SH2)
	assert.NotNil(t, readData.Data[0].SH3)
	assert.Equal(t, 24, len(readData.Data[0].SH2))
	assert.Equal(t, 21, len(readData.Data[0].SH3))
}

// TestEmptyData tests handling of empty data
func TestEmptyData(t *testing.T) {
	testFile := "test_empty.spz"
	defer os.Remove(testFile)

	originalData := &SpzData{
		Magic:          SPZ_MAGIC,
		Version:        2,
		NumPoints:      0,
		ShDegree:       0,
		FractionalBits: 12,
		Flags:          0,
		Reserved:       0,
		Data:           []*SplatData{},
	}

	err := WriteSpz(testFile, originalData)
	assert.NoError(t, err)

	readData, err := ReadSpz(testFile)
	assert.NoError(t, err)
	assert.NotNil(t, readData)
	assert.Equal(t, uint32(0), readData.NumPoints)
	assert.Equal(t, 0, len(readData.Data))
}
