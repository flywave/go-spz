package spz

import (
	"os"
)

// WriteSpz writes SPZ data to file
func WriteSpz(spzFile string, spzData *SpzData) error {
	file, err := os.Create(spzFile)
	if err != nil {
		return err
	}
	defer file.Close()

	bts := make([]byte, 0)
	bts = append(bts, spzData.ToBytes()...)

	rows := spzData.Data

	// Encode positions
	for i := range rows {
		bts = append(bts, spzEncodePosition(rows[i].PositionX)...)
		bts = append(bts, spzEncodePosition(rows[i].PositionY)...)
		bts = append(bts, spzEncodePosition(rows[i].PositionZ)...)
	}

	// Encode alphas
	for i := range rows {
		bts = append(bts, rows[i].ColorA)
	}

	// Encode colors
	for i := range rows {
		bts = append(bts, spzEncodeColor(rows[i].ColorR), spzEncodeColor(rows[i].ColorG), spzEncodeColor(rows[i].ColorB))
	}

	// Encode scales
	for i := range rows {
		bts = append(bts, spzEncodeScale(rows[i].ScaleX), spzEncodeScale(rows[i].ScaleY), spzEncodeScale(rows[i].ScaleZ))
	}

	// Encode rotations
	for i := range rows {
		if spzData.Version >= 3 {
			bts = append(bts, spzEncodeRotationsV3(rows[i].RotationW, rows[i].RotationX, rows[i].RotationY, rows[i].RotationZ)...)
		} else {
			bts = append(bts, spzEncodeRotations(rows[i].RotationW, rows[i].RotationX, rows[i].RotationY, rows[i].RotationZ)...)
		}
	}

	// Encode SH data
	switch spzData.ShDegree {
	case 1:
		for i := range rows {
			if len(rows[i].SH1) > 0 {
				for n := range 9 {
					bts = append(bts, spzEncodeSH1(rows[i].SH1[n]))
				}
			} else if len(rows[i].SH2) > 0 {
				for n := range 9 {
					bts = append(bts, spzEncodeSH1(rows[i].SH2[n]))
				}
			} else {
				for range 9 {
					bts = append(bts, encodeSplatSH(0.0))
				}
			}
		}
	case 2:
		for i := range rows {
			if len(rows[i].SH1) > 0 {
				for n := range 9 {
					bts = append(bts, spzEncodeSH1(rows[i].SH1[n]))
				}
				for range 15 {
					bts = append(bts, encodeSplatSH(0.0))
				}
			} else if len(rows[i].SH2) > 0 {
				for n := range 9 {
					bts = append(bts, spzEncodeSH1(rows[i].SH2[n]))
				}
				for j := 9; j < 24; j++ {
					bts = append(bts, spzEncodeSH23(rows[i].SH2[j]))
				}
			} else {
				for range 24 {
					bts = append(bts, encodeSplatSH(0.0))
				}
			}
		}
	case 3:
		for i := range rows {
			if len(rows[i].SH3) > 0 {
				for j := range 9 {
					bts = append(bts, spzEncodeSH1(rows[i].SH2[j]))
				}
				for j := 9; j < 24; j++ {
					bts = append(bts, spzEncodeSH23(rows[i].SH2[j]))
				}
				for j := range 21 {
					bts = append(bts, spzEncodeSH23(rows[i].SH3[j]))
				}
			} else if len(rows[i].SH2) > 0 {
				for j := range 9 {
					bts = append(bts, spzEncodeSH1(rows[i].SH2[j]))
				}
				for j := 9; j < 24; j++ {
					bts = append(bts, spzEncodeSH23(rows[i].SH2[j]))
				}
				for range 21 {
					bts = append(bts, encodeSplatSH(0.0))
				}
			} else if len(rows[i].SH1) > 0 {
				for n := range 9 {
					bts = append(bts, spzEncodeSH1(rows[i].SH1[n]))
				}
				for range 36 {
					bts = append(bts, encodeSplatSH(0.0))
				}
			} else {
				for range 45 {
					bts = append(bts, encodeSplatSH(0.0))
				}
			}
		}
	}

	// Compress with gzip
	gzipDatas, err := compressGzip(bts)
	if err != nil {
		return err
	}

	// Write to file
	_, err = file.Write(gzipDatas)
	if err != nil {
		return err
	}

	return nil
}
