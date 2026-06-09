package civ6save

import (
	"encoding/binary"
	"errors"
)

type reader struct {
	data []byte
	pos  int
}

func newReader(data []byte, pos int) *reader {
	return &reader{data: data, pos: pos}
}

func (r *reader) remaining() int { return len(r.data) - r.pos }

func (r *reader) readU8() uint8 {
	v := r.data[r.pos]
	r.pos++
	return v
}

func (r *reader) readU16() uint16 {
	v := binary.LittleEndian.Uint16(r.data[r.pos:])
	r.pos += 2
	return v
}

func (r *reader) readU32() uint32 {
	v := binary.LittleEndian.Uint32(r.data[r.pos:])
	r.pos += 4
	return v
}

func (r *reader) readI32() int {
	return int(int32(r.readU32()))
}

// readInt mirrors buffToInteger(it, byteCount) — reads 1,2,3,4 bytes LE.
func (r *reader) readInt(byteCount int) uint32 {
	var v uint32
	for i := 0; i < byteCount; i++ {
		v |= uint32(r.data[r.pos+i]) << (8 * i)
	}
	r.pos += byteCount
	return v
}

func (r *reader) skip(n int) {
	r.pos += n
}

func (r *reader) peek32() uint32 {
	return binary.LittleEndian.Uint32(r.data[r.pos:])
}

// readFloat mirrors readFloat in C++: value>>8 + (value&0xff)/256.
func (r *reader) readFloat() float32 {
	v := r.readU32()
	a := int(v >> 8)
	b := v & 0xff
	return float32(a) + float32(b)/256.0
}

// readString mirrors readString(it, sizeCount=4).
func (r *reader) readString() string {
	count := int(r.readU32())
	if count == 0 {
		return ""
	}
	s := string(r.data[r.pos : r.pos+count])
	r.pos += count
	return s
}

// readStringN mirrors readString(it, sizeCount=N).
func (r *reader) readStringN(sizeCount int) string {
	count := int(r.readInt(sizeCount))
	if count == 0 {
		return ""
	}
	s := string(r.data[r.pos : r.pos+count])
	r.pos += count
	return s
}

// readMap mirrors readMap(it, sizeValue=4) — returns map[key]value.
func (r *reader) readMap(sizeValue int) map[uint32]uint32 {
	count := int(r.readU32())
	m := make(map[uint32]uint32, count)
	for i := 0; i < count; i++ {
		key := r.readU32()
		val := r.readInt(sizeValue)
		m[key] = val
	}
	return m
}

// readMapDiscard mirrors readMap(it) with no output.
func (r *reader) readMapDiscard() {
	count := int(r.readU32())
	r.skip(count * 8)
}

// readMapDiscardSV mirrors readMap(it, sizeValue).
func (r *reader) readMapDiscardSV(sizeValue int) {
	count := int(r.readU32())
	r.skip(count * (4 + sizeValue))
}

// readMapFloat mirrors readMapFloat(it).
func (r *reader) readMapFloat() map[uint32]float32 {
	count := int(r.readU32())
	m := make(map[uint32]float32, count)
	for i := 0; i < count; i++ {
		key := r.readU32()
		m[key] = r.readFloat()
	}
	return m
}

// readMapBool mirrors readMapBool(it).
func (r *reader) readMapBoolDiscard() {
	count := int(r.readU32())
	r.skip(count * 5) // key(4) + bool(1)
}

// readMapBool reads a bool map and returns a set of keys where value is true.
func (r *reader) readMapBool() map[uint32]bool {
	count := int(r.readU32())
	m := make(map[uint32]bool, count)
	for i := 0; i < count; i++ {
		key := r.readU32()
		val := r.readU8()
		if val != 0 {
			m[key] = true
		}
	}
	return m
}

// readArrayDiscard mirrors readArray(it, size, sep, sizeCount).
// sep=1 means skip 1 byte after non-zero values.
func (r *reader) readArrayDiscard(size int, sep int) {
	count := int(r.readU32())
	for i := 0; i < count; i++ {
		v := r.readInt(size)
		if v != 0 && sep > 0 {
			r.skip(sep)
		}
	}
}

// readArrayDiscardSC mirrors readArray with custom sizeCount.
func (r *reader) readArrayDiscardSC(size int, sep int, sizeCount int) {
	count := int(r.readInt(sizeCount))
	for i := 0; i < count; i++ {
		v := r.readInt(size)
		if v != 0 && sep > 0 {
			r.skip(sep)
		}
	}
}

func (r *reader) assertU32(expected uint32) error {
	got := r.peek32()
	if got != expected {
		return errors.New("assertion failed")
	}
	return nil
}
