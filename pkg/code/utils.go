package code

import "encoding/binary"

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
