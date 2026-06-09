package civ6save

import (
	"bytes"
	"compress/zlib"
	"errors"
	"io"
)

var modTitle = []byte("MOD_TITLE")
var zlibMagic = []byte{0x78, 0x9C}
var endMarker = []byte{0x00, 0x00, 0xFF, 0xFF, 0x02, 0x00, 0x00, 0x00}

func Decompress(data []byte) ([]byte, error) {
	modIdx := bytes.LastIndex(data, modTitle)
	if modIdx == -1 {
		return nil, errors.New("MOD_TITLE not found")
	}

	zlibIdx := bytes.Index(data[modIdx:], zlibMagic)
	if zlibIdx == -1 {
		return nil, errors.New("zlib magic not found")
	}
	zlibIdx += modIdx

	endIdx := bytes.Index(data[zlibIdx:], endMarker)
	if endIdx == -1 {
		return nil, errors.New("end marker not found")
	}
	endIdx += zlibIdx

	raw := data[zlibIdx:endIdx]

	const chunkSize = 64 * 1024
	var compressed []byte
	pos := 0
	for pos < len(raw) {
		end := pos + chunkSize
		if end > len(raw) {
			end = len(raw)
		}
		compressed = append(compressed, raw[pos:end]...)
		pos = end
		if pos < len(raw) {
			pos += 4
		}
	}

	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var out bytes.Buffer
	_, err = io.Copy(&out, r)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, err
	}
	return out.Bytes(), nil
}
