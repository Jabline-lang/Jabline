package code

func ReadUint16(ins Instructions) uint16 {
	return uint16(ins[0])<<8 | uint16(ins[1])
}

func WriteUint16(ins Instructions, offset int, val uint16) {
	ins[offset] = byte(val >> 8)
	ins[offset+1] = byte(val)
}
