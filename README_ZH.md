# go-spz

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

中文 | [English](README.md)

一个用于读写 SPZ (Niantic Gaussian Splat) 文件格式的 Go 语言库。

## 简介

SPZ 是一种用于存储 3D Gaussian Splatting 数据的二进制文件格式。该库提供了完整的 SPZ 文件读写功能，支持：

- **多版本支持**: Version 2 和 Version 3 格式
- **球谐函数**: 支持 SH Degree 0/1/2/3
- **数据压缩**: 自动 Gzip 压缩/解压
- **高效编码**: 优化的数据编码方案
  - 位置: 24-bit 定点数
  - 缩放: 8-bit 量化
  - 旋转: Version 2 (3 bytes), Version 3 (4 bytes)
  - 颜色: SH-based 编码

## 安装

```bash
go get github.com/flywave/go-spz
```

## 快速开始

### 读取 SPZ 文件

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/flywave/go-spz"
)

func main() {
    // 读取 SPZ 文件
    data, err := spz.ReadSpz("input.spz")
    if err != nil {
        log.Fatal(err)
    }
    
    // 访问头部信息
    fmt.Printf("Version: %d\n", data.Version)
    fmt.Printf("Points: %d\n", data.NumPoints)
    fmt.Printf("SH Degree: %d\n", data.ShDegree)
    
    // 访问点数据
    for i, point := range data.Data {
        fmt.Printf("Point %d: Position(%.2f, %.2f, %.2f)\n", 
            i, point.PositionX, point.PositionY, point.PositionZ)
    }
}
```

### 写入 SPZ 文件

```go
package main

import (
    "log"
    
    "github.com/flywave/go-spz"
)

func main() {
    // 创建 SPZ 数据
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
    
    // 写入文件
    err := spz.WriteSpz("output.spz", data)
    if err != nil {
        log.Fatal(err)
    }
}
```

## 文件格式

### SPZ 文件结构

```
┌─────────────────────────────────────────┐
│          GZIP 压缩数据                   │
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

### 版本差异

| 特性 | Version 2 | Version 3 |
|------|-----------|-----------|
| 旋转编码 | 3 bytes (重建第4个分量) | 4 bytes (高精度) |
| 精度 | 标准 | 更高 |
| 文件大小 | 更小 | 稍大 |

### SH Degree 数据量

| Degree | 每点字节数 | 说明 |
|--------|-----------|------|
| 0 | 0 | 无 SH 数据 |
| 1 | 9 | SH1 (9 bytes) |
| 2 | 24 | SH2 (24 bytes) |
| 3 | 45 | SH2 (24 bytes) + SH3 (21 bytes) |

## API 文档

### 主要类型

#### SpzData
```go
type SpzData struct {
    Magic          uint32        // 文件魔数
    Version        uint32        // 版本号 (2 或 3)
    NumPoints      uint32        // 点的数量
    ShDegree       uint8         // 球谐函数阶数 (0-3)
    FractionalBits uint8         // 定点数精度位数
    Flags          uint8         // 标志位
    Reserved       uint8         // 保留字段
    Data           []*SplatData  // 点数据
}
```

#### SplatData
```go
type SplatData struct {
    PositionX, PositionY, PositionZ float32  // 位置
    ScaleX, ScaleY, ScaleZ          float32  // 缩放
    RotationW, RotationX, RotationY, RotationZ uint8  // 旋转（四元数）
    ColorR, ColorG, ColorB, ColorA  uint8    // 颜色
    SH1, SH2, SH3                   []byte   // 球谐函数数据
}
```

### 主要函数

#### ReadSpz
```go
func ReadSpz(file string) (*SpzData, error)
```
读取 SPZ 文件并返回解析后的数据。

**参数:**
- `file`: 文件路径

**返回:**
- `*SpzData`: 解析后的 SPZ 数据
- `error`: 错误信息

#### WriteSpz
```go
func WriteSpz(spzFile string, spzData *SpzData) error
```
将 SPZ 数据写入文件。

**参数:**
- `spzFile`: 输出文件路径
- `spzData`: 要写入的 SPZ 数据

**返回:**
- `error`: 错误信息

#### ParseSpzHeader
```go
func ParseSpzHeader(data []byte) (*SpzData, error)
```
解析 SPZ 文件头部。

#### ToBytes
```go
func (h *SpzData) ToBytes() []byte
```
将 SPZ 头部序列化为字节数组。

## 编码说明

### 位置编码
- 使用 24-bit 定点数
- 精度: 1/4096
- 范围: 约 ±2048

### 缩放编码
```
编码: (scale + 10.0) × 16
解码: value / 16.0 - 10.0
范围: -10.0 ~ 5.9375
```

### 颜色编码
- 基于球谐函数的编码
- 有损压缩以节省空间

### 旋转编码

**Version 2:**
- 存储 3 个分量
- 通过四元数单位约束重建第 4 个分量

**Version 3:**
- 存储最大分量索引 (2 bits)
- 存储其他 3 个分量 (30 bits)
- 更高精度

## 测试

运行测试：

```bash
go test -v
```

运行特定测试：

```bash
go test -v -run TestReadWriteSpz
```

查看测试覆盖率：

```bash
go test -cover
```

## 项目结构

```
go-spz/
├── spz.go              # 核心数据结构定义
├── header.go           # 头部序列化/反序列化
├── read.go             # 文件读取功能
├── write.go            # 文件写入功能
├── utils.go            # 编码/解码工具函数
├── gzip.go             # Gzip 压缩/解压
├── error.go            # 错误定义
├── spz_test.go         # 单元测试
└── README.md           # 本文件
```

## 性能考虑

- **内存效率**: 使用流式处理，避免大量内存分配
- **压缩**: 自动 Gzip 压缩，通常可减少 70-80% 文件大小
- **编码优化**: 针对 3D 点云数据特点优化的编码方案

## 注意事项

1. **有损编码**: 颜色和旋转数据使用有损编码，存在精度损失
2. **版本兼容**: Version 2 和 3 不完全兼容，注意版本选择
3. **SH 数据**: 确保 SH 数据长度与 ShDegree 匹配
4. **四元数**: 旋转数据应为有效的单位四元数

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 相关链接

- [3D Gaussian Splatting](https://repo-sam.inria.fr/fungraph/3d-gaussian-splatting/)
- [Go Documentation](https://golang.org/doc/)
