# go-spz

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

[中文](README_ZH.md) | English

A Go library for reading and writing SPZ (Niantic Gaussian Splat) file format.

## Introduction

SPZ is a binary file format for storing 3D Gaussian Splatting data. This library provides complete SPZ file read/write functionality, supporting:

- **Multi-version Support**: Version 2 and Version 3 formats
- **Spherical Harmonics**: Support for SH Degree 0/1/2/3
- **Data Compression**: Automatic Gzip compression/decompression
- **Efficient Encoding**: Optimized data encoding scheme
  - Position: 24-bit fixed-point
  - Scale: 8-bit quantization
  - Rotation: Version 2 (3 bytes), Version 3 (4 bytes)
  - Color: SH-based encoding

## Installation

```bash
go get github.com/flywave/go-spz
```

## Quick Start

### Reading SPZ Files

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/flywave/go-spz"
)

func main() {
    // Read SPZ file
    data, err := spz.ReadSpz("input.spz")
    if err != nil {
        log.Fatal(err)
    }
    
    // Access header information
    fmt.Printf("Version: %d\n", data.Version)
    fmt.Printf("Points: %d\n", data.NumPoints)
    fmt.Printf("SH Degree: %d\n", data.ShDegree)
    
    // Access point data
    for i, point := range data.Data {
        fmt.Printf("Point %d: Position(%.2f, %.2f, %.2f)\n", 
            i, point.PositionX, point.PositionY, point.PositionZ)
    }
}
```

### Writing SPZ Files

```go
package main

import (
    "log"
    
    "github.com/flywave/go-spz"
)

func main() {
    // Create SPZ data
    data := &spz.SpzData{
        Magic:          spz.SPZ_MAGIC,
        Version:        2,
        NumPoints:      2,
        ShDegree:       1,
        FractionalBits: 12,
        Flags:          0,
        Reserved:       0,
        Data: []*spz.SplatData{
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
        },
    }
    
    // Write to file
    err := spz.WriteSpz("output.spz", data)
    if err != nil {
        log.Fatal(err)
    }
}
```

## File Format

### SPZ File Structure

```
┌─────────────────────────────────────────┐
│          GZIP Compressed Data           │
├─────────────────────────────────────────┤
│  Header (16 bytes)                      │
│  ├─ Magic (4 bytes): 0x5053474E         │
│  ├─ Version (4 bytes): 2 or 3           │
│  ├─ NumPoints (4 bytes)                 │
│  ├─ ShDegree (1 byte): 0-3              │
│  ├─ FractionalBits (1 byte): 12         │
│  ├─ Flags (1 byte)                      │
│  └─ Reserved (1 byte)                   │
├─────────────────────────────────────────┤
│  Data Section                           │
│  ├─ Positions (9 bytes × N)             │
│  ├─ Alphas (1 byte × N)                 │
│  ├─ Colors (3 bytes × N)                │
│  ├─ Scales (3 bytes × N)                │
│  ├─ Rotations (3/4 bytes × N)           │
│  └─ SH Data (varies by degree)          │
└─────────────────────────────────────────┘
```

### Version Differences

| Feature | Version 2 | Version 3 |
|---------|-----------|-----------|
| Rotation Encoding | 3 bytes (rebuild 4th component) | 4 bytes (higher precision) |
| Precision | Standard | Higher |
| File Size | Smaller | Slightly larger |

### SH Degree Data Size

| Degree | Bytes per Point | Description |
|--------|-----------------|-------------|
| 0 | 0 | No SH data |
| 1 | 9 | SH1 (9 bytes) |
| 2 | 24 | SH2 (24 bytes) |
| 3 | 45 | SH2 (24 bytes) + SH3 (21 bytes) |

## API Documentation

### Main Types

#### SpzData
```go
type SpzData struct {
    Magic          uint32        // File magic number
    Version        uint32        // Version number (2 or 3)
    NumPoints      uint32        // Number of points
    ShDegree       uint8         // Spherical harmonics degree (0-3)
    FractionalBits uint8         // Fixed-point precision bits
    Flags          uint8         // Flags
    Reserved       uint8         // Reserved field
    Data           []*SplatData  // Point data
}
```

#### SplatData
```go
type SplatData struct {
    PositionX, PositionY, PositionZ float32  // Position
    ScaleX, ScaleY, ScaleZ          float32  // Scale
    RotationW, RotationX, RotationY, RotationZ uint8  // Rotation (quaternion)
    ColorR, ColorG, ColorB, ColorA  uint8    // Color
    SH1, SH2, SH3                   []byte   // Spherical harmonics data
}
```

### Main Functions

#### ReadSpz
```go
func ReadSpz(file string) (*SpzData, error)
```
Reads an SPZ file and returns the parsed data.

**Parameters:**
- `file`: File path

**Returns:**
- `*SpzData`: Parsed SPZ data
- `error`: Error information

#### WriteSpz
```go
func WriteSpz(spzFile string, spzData *SpzData) error
```
Writes SPZ data to a file.

**Parameters:**
- `spzFile`: Output file path
- `spzData`: SPZ data to write

**Returns:**
- `error`: Error information

#### ParseSpzHeader
```go
func ParseSpzHeader(data []byte) (*SpzData, error)
```
Parses the SPZ file header.

#### ToBytes
```go
func (h *SpzData) ToBytes() []byte
```
Serializes the SPZ header to a byte array.

## Encoding Details

### Position Encoding
- Uses 24-bit fixed-point numbers
- Precision: 1/4096
- Range: approximately ±2048

### Scale Encoding
```
Encode: (scale + 10.0) × 16
Decode: value / 16.0 - 10.0
Range: -10.0 ~ 5.9375
```

### Color Encoding
- Spherical harmonics-based encoding
- Lossy compression to save space

### Rotation Encoding

**Version 2:**
- Stores 3 components
- Reconstructs the 4th component via unit quaternion constraint

**Version 3:**
- Stores maximum component index (2 bits)
- Stores other 3 components (30 bits)
- Higher precision

## Testing

Run tests:

```bash
go test -v
```

Run specific tests:

```bash
go test -v -run TestReadWriteSpz
```

View test coverage:

```bash
go test -cover
```

## Project Structure

```
go-spz/
├── spz.go              # Core data structure definitions
├── header.go           # Header serialization/deserialization
├── read.go             # File reading functionality
├── write.go            # File writing functionality
├── utils.go            # Encoding/decoding utility functions
├── gzip.go             # Gzip compression/decompression
├── error.go            # Error definitions
├── spz_test.go         # Unit tests
├── README.md           # Chinese documentation
└── README_EN.md        # This file
```

## Performance Considerations

- **Memory Efficiency**: Uses streaming processing to avoid excessive memory allocation
- **Compression**: Automatic Gzip compression, typically reduces file size by 70-80%
- **Encoding Optimization**: Encoding scheme optimized for 3D point cloud data characteristics

## Notes

1. **Lossy Encoding**: Color and rotation data use lossy encoding with precision loss
2. **Version Compatibility**: Version 2 and 3 are not fully compatible, pay attention to version selection
3. **SH Data**: Ensure SH data length matches ShDegree
4. **Quaternions**: Rotation data should be valid unit quaternions

## License

MIT License

## Contributing

Issues and Pull Requests are welcome!

## Related Links

- [3D Gaussian Splatting](https://repo-sam.inria.fr/fungraph/3d-gaussian-splatting/)
- [Go Documentation](https://golang.org/doc/)
