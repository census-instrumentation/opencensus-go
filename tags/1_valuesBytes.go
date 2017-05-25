package tags

import "unsafe"

var sizeOfUint16 = (int)(unsafe.Sizeof(uint16(0)))

type valueBytes struct {
	bytes []byte
	wIdx, rIdx int
}

func (vb *valueBytes) grow(s int) {
	for ;len(vb.bytes) <= s; {
		newSize := 2*len(vb.bytes)+1
		tmp := make([]byte, newSize)
		copy(tmp, vb.bytes)
		vb.bytes = tmp
	}
}

func (vb *valueBytes) writeValue(v []byte) {
	length := len(v)
	endIdx := vb.wIdx + sizeOfUint16 + int(length)
	vb.grow(endIdx)
	
	// writing length of v
	bytes := *(*[2]byte)(unsafe.Pointer(&length))
	vb.bytes[vb.wIdx] = bytes[0]
	vb.wIdx++
	vb.bytes[vb.wIdx] = bytes[1]
	vb.wIdx++

	if length == 0 {
		// No value was encoded for this key
		return
	}

	// writing v
	copy(vb.bytes[vb.wIdx:], v)
	vb.wIdx = endIdx
}

func (vb *valueBytes) readValue() []byte {
	// read length of v
	length := (int)(*(*uint16)(unsafe.Pointer(&vb.bytes[vb.rIdx])))
	vb.rIdx += sizeOfUint16
	if length == 0 {
		// No value was encoded for this key
		return nil
	}

	// read value of v
	v := make([]byte, length)
	endIdx := vb.rIdx+length
	copy(v, vb.bytes[vb.rIdx:endIdx])
	vb.rIdx += endIdx
	return v
}


func (vb *valueBytes) toMap(ks []Key) map[Key][]byte {
	m := make(map[Key][]byte, len(ks))
	for _, k := range ks {
		v := vb.readValue()
		if v != nil {
			m[k] = v
		}
	}
	return m	
}