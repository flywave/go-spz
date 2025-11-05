package spz

import (
	"io"
	"os"
)

// readSpzDatas parses the data section of an SPZ file
func readSpzDatas(datas []byte, h *SpzData) ([]*SplatData, error) {
	// Calculate sizes for each data section
	positionSize := int(h.NumPoints) * 9 // 3 bytes per axis * 3 axes
	alphaSize := int(h.NumPoints)        // 1 byte per point
	colorSize := int(h.NumPoints) * 3    // 1 byte per channel * 3 channels
	scaleSize := int(h.NumPoints) * 3    // 1 byte per axis * 3 axes
	rotationSize := int(h.NumPoints) * 3 // Version 2: 3 bytes
	if h.Version >= 3 {
		rotationSize = int(h.NumPoints) * 4 // Version 3: 4 bytes
	}

	// Calculate SH data size
	shDim := 0
	switch h.ShDegree {
	case 1:
		shDim = int(h.NumPoints) * 9
	case 2:
		shDim = int(h.NumPoints) * 24
	case 3:
		shDim = int(h.NumPoints) * 45
	}

	// Calculate expected data size
	expectedSize := positionSize + alphaSize + colorSize + scaleSize + rotationSize + shDim

	// Validate data size
	if len(datas) != expectedSize {
		return nil, &SpzError{"Invalid SPZ data: incorrect data size"}
	}

	// Calculate offsets (matching write order)
	offsetPositions := 0
	offsetAlphas := offsetPositions + positionSize
	offsetColors := offsetAlphas + alphaSize
	offsetScales := offsetColors + colorSize
	offsetRotations := offsetScales + scaleSize
	offsetShs := offsetRotations + rotationSize

	// Extract data sections
	positions := datas[offsetPositions:offsetAlphas]
	alphas := datas[offsetAlphas:offsetColors]
	colors := datas[offsetColors:offsetScales]
	scales := datas[offsetScales:offsetRotations]
	rotations := datas[offsetRotations:offsetShs]
	shs := datas[offsetShs:]

	// Parse each splat data point
	var splatDatas []*SplatData
	for i := range int(h.NumPoints) {
		data := &SplatData{}

		// Decode positions (3 bytes each)
		data.PositionX = spzDecodePosition(positions[i*9:i*9+3], h.FractionalBits)
		data.PositionY = spzDecodePosition(positions[i*9+3:i*9+6], h.FractionalBits)
		data.PositionZ = spzDecodePosition(positions[i*9+6:i*9+9], h.FractionalBits)

		// Decode alpha (1 byte)
		data.ColorA = alphas[i]

		// Decode colors (1 byte each, with decoding)
		data.ColorR = spzDecodeColor(colors[i*3])
		data.ColorG = spzDecodeColor(colors[i*3+1])
		data.ColorB = spzDecodeColor(colors[i*3+2])

		// Decode scales (1 byte each)
		data.ScaleX = spzDecodeScale(scales[i*3])
		data.ScaleY = spzDecodeScale(scales[i*3+1])
		data.ScaleZ = spzDecodeScale(scales[i*3+2])

		// Decode rotations (version dependent)
		if h.Version >= 3 {
			data.RotationW, data.RotationX, data.RotationY, data.RotationZ = spzDecodeRotationsV3(rotations[i*4 : i*4+4])
		} else {
			data.RotationW, data.RotationX, data.RotationY, data.RotationZ = spzDecodeRotations(rotations[i*3], rotations[i*3+1], rotations[i*3+2])
		}

		// Decode SH data (if present)
		switch h.ShDegree {
		case 1:
			if i*9+9 <= len(shs) {
				data.SH1 = make([]byte, 9)
				for j := 0; j < 9; j++ {
					data.SH1[j] = spzDecodeSH1(shs[i*9+j])
				}
			}
		case 2:
			if i*24+24 <= len(shs) {
				data.SH2 = make([]byte, 24)
				for j := 0; j < 9; j++ {
					data.SH2[j] = spzDecodeSH1(shs[i*24+j])
				}
				for j := 9; j < 24; j++ {
					data.SH2[j] = spzDecodeSH23(shs[i*24+j])
				}
			}
		case 3:
			if i*45+45 <= len(shs) {
				data.SH2 = make([]byte, 24)
				for j := 0; j < 9; j++ {
					data.SH2[j] = spzDecodeSH1(shs[i*45+j])
				}
				for j := 9; j < 24; j++ {
					data.SH2[j] = spzDecodeSH23(shs[i*45+j])
				}

				data.SH3 = make([]byte, 21)
				for j := 0; j < 21; j++ {
					data.SH3[j] = spzDecodeSH23(shs[i*45+24+j])
				}
			}
		}

		splatDatas = append(splatDatas, data)
	}

	return splatDatas, nil
}

// ReadSpz reads an SPZ file and returns its header and data
func ReadSpz(file string) (*SpzData, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read all data
	gzipDatas, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// Decompress gzip data
	ungzipDatas, err := decompressGzip(gzipDatas)
	if err != nil {
		// If decompression fails, assume the data is not compressed
		ungzipDatas = gzipDatas
	}

	// Check if we have enough data for the header
	if len(ungzipDatas) < HeaderSizeSpz {
		return nil, &SpzError{"Invalid SPZ file: insufficient data for header"}
	}

	// Parse header
	spzData, err := ParseSpzHeader(ungzipDatas[0:HeaderSizeSpz])
	if err != nil {
		return nil, err
	}
	if spzData == nil {
		return nil, &SpzError{"Failed to parse SPZ header"}
	}

	// Parse data
	if len(ungzipDatas) > HeaderSizeSpz {
		spzData.Data, err = readSpzDatas(ungzipDatas[HeaderSizeSpz:], spzData)
		if err != nil {
			return nil, err
		}
	} else {
		spzData.Data = []*SplatData{}
	}

	return spzData, nil
}
