// Package fontutil loads the system CJK font Godot falls back to.
//
// Microsoft YaHei ships as msyh.ttc, a TrueType *collection*: several faces
// sharing one pool of tables. The parser underneath g3n (golang/freetype)
// only understands a plain sfnt file, so ExtractFace rebuilds face 0 of a
// collection into a standalone in-memory font. Nothing is written to disk and
// no font is redistributed -- the bytes come from the machine's own font
// directory at startup.
package fontutil

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/g3n/engine/text"
)

// Load returns the first candidate that parses, and the path it came from.
func Load(candidates []string, pointSize float64) (*text.Font, string, error) {
	var lastErr error
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			lastErr = err
			continue
		}
		if IsCollection(data) {
			data, err = ExtractFace(data, 0)
			if err != nil {
				lastErr = fmt.Errorf("%s: %w", path, err)
				continue
			}
		}
		font, err := text.NewFontFromData(data)
		if err != nil {
			lastErr = fmt.Errorf("%s: %w", path, err)
			continue
		}
		font.SetPointSize(pointSize)
		return font, path, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no candidates given")
	}
	return nil, "", lastErr
}

// IsCollection reports whether data starts with the 'ttcf' tag.
func IsCollection(data []byte) bool {
	return len(data) >= 4 && string(data[0:4]) == "ttcf"
}

// ExtractFace rebuilds face `index` of a TrueType collection as a standalone
// sfnt file.
//
// A collection is an ordinary sfnt table pool with several table directories
// pointing into it, so extracting a face means writing a fresh header and
// directory whose records point at copies of just that face's tables.
func ExtractFace(data []byte, index int) ([]byte, error) {
	if !IsCollection(data) {
		return nil, fmt.Errorf("not a font collection")
	}
	if len(data) < 12 {
		return nil, fmt.Errorf("truncated collection header")
	}

	numFonts := int(binary.BigEndian.Uint32(data[8:12]))
	if index < 0 || index >= numFonts {
		return nil, fmt.Errorf("face %d out of range (collection has %d)", index, numFonts)
	}
	offsetPos := 12 + 4*index
	if len(data) < offsetPos+4 {
		return nil, fmt.Errorf("truncated collection offset table")
	}
	dirOffset := int(binary.BigEndian.Uint32(data[offsetPos : offsetPos+4]))
	if len(data) < dirOffset+12 {
		return nil, fmt.Errorf("truncated table directory")
	}

	sfntVersion := binary.BigEndian.Uint32(data[dirOffset : dirOffset+4])
	numTables := int(binary.BigEndian.Uint16(data[dirOffset+4 : dirOffset+6]))
	if len(data) < dirOffset+12+16*numTables {
		return nil, fmt.Errorf("truncated table records")
	}

	type table struct {
		tag      [4]byte
		checksum uint32
		data     []byte
	}
	tables := make([]table, 0, numTables)
	for i := 0; i < numTables; i++ {
		rec := data[dirOffset+12+16*i:]
		var t table
		copy(t.tag[:], rec[0:4])
		t.checksum = binary.BigEndian.Uint32(rec[4:8])
		off := int(binary.BigEndian.Uint32(rec[8:12]))
		length := int(binary.BigEndian.Uint32(rec[12:16]))
		if off < 0 || length < 0 || off+length > len(data) {
			return nil, fmt.Errorf("table %q lies outside the file", t.tag)
		}
		t.data = data[off : off+length]
		tables = append(tables, t)
	}

	// Header, then the directory, then the tables, each padded to 4 bytes.
	headerSize := 12 + 16*numTables
	out := make([]byte, headerSize, headerSize+len(data)/2)

	binary.BigEndian.PutUint32(out[0:4], sfntVersion)
	binary.BigEndian.PutUint16(out[4:6], uint16(numTables))
	// searchRange / entrySelector / rangeShift: derived, and no parser that
	// matters relies on them, but they are cheap to get right.
	entrySelector := uint16(0)
	for 1<<(entrySelector+1) <= uint16(numTables) {
		entrySelector++
	}
	searchRange := uint16(16) << entrySelector
	binary.BigEndian.PutUint16(out[6:8], searchRange)
	binary.BigEndian.PutUint16(out[8:10], entrySelector)
	binary.BigEndian.PutUint16(out[10:12], uint16(16*numTables)-searchRange)

	for i, t := range tables {
		offset := len(out)
		out = append(out, t.data...)
		for len(out)%4 != 0 {
			out = append(out, 0)
		}
		rec := out[12+16*i:]
		copy(rec[0:4], t.tag[:])
		binary.BigEndian.PutUint32(rec[4:8], t.checksum)
		binary.BigEndian.PutUint32(rec[8:12], uint32(offset))
		binary.BigEndian.PutUint32(rec[12:16], uint32(len(t.data)))
	}

	return out, nil
}
