package tcpsocket

type PulsePucket struct {
	opcode  uint16
	size    uint32
	content []byte
}

func (p *PulsePucket) ToSlice() []byte {
	oBuf := make([]byte, 2)
	sBuf := make([]byte, 4)

	// move uint16 to buffer
	oBuf[0] = byte((p.opcode >> 8) & 0xFF)
	oBuf[1] = byte(p.opcode & 0xFF)

	// move uint32 to buffer
	sBuf[0] = byte((p.size >> 24) & 0xFF)
	sBuf[1] = byte((p.size >> 16) & 0xFF)
	sBuf[2] = byte((p.size >> 8) & 0xFF)
	sBuf[3] = byte(p.size & 0xFF)

	result := append(oBuf, sBuf...)
	result = append(result, p.content...)

	return result
}

func InitPacket(opcode uint16, content []byte) PulsePucket {
	p := PulsePucket{}
	p.content = content
	p.opcode = opcode
	p.size = uint32(len(content))
	return p
}

func InitEmptyPacket(opcode uint16) PulsePucket {
	p := PulsePucket{}
	p.content = make([]byte, 0)
	p.opcode = opcode
	p.size = uint32(len(p.content))
	return p
}

func InitPacketWithHeaderData(header []byte) PulsePucket {
	p := PulsePucket{}

	p.opcode = uint16(header[0])<<8 | uint16(header[1])
	p.size = uint32(header[2])<<24 | uint32(header[3])<<16 | uint32(header[4])<<8 | uint32(header[5])

	p.content = make([]byte, 0)
	return p
}
