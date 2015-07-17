package tcpsocket

import (
	//"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/walkline/MoC-pulse-backend/auth"
	"log"
	"net"
)

const (
	CS_AUTH = iota
	SC_AUTH
	CS_GET_VOTES
	SC_GET_VOTES_RESULT
	CS_GET_VOTE
	SC_GET_VOTE_RESULT
	CS_VOTE_FOR
	SC_UPDATE_VOTE
	CS_CREATE_VOTE
	SC_NEW_VOTE
)

type TcpConn struct {
	user      auth.User
	conection *net.Conn
}

type PulsePucket struct {
	opcode  uint16
	size    uint32
	content []byte
}

func (c *TcpConn) SendPacket(p *PulsePucket) {
	(*(*c).conection).Write(p.ToSlice())
}

func (p *PulsePucket) ToSlice() []byte {
	oBuf := make([]byte, 2)
	sBuf := make([]byte, 4)
	binary.LittleEndian.PutUint16(oBuf, p.opcode)
	binary.LittleEndian.PutUint32(sBuf, p.size)

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

var globalWriteChan = make(chan *PulsePucket)

func HandleNewConnection(c net.Conn) {
	user := auth.User{
		Id:     "debug",
		Email:  "test@test.com",
		Device: 2,
		DevId:  "",
	}

	tcpConnection := TcpConn{user, &c}

	// json to []byte
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(user)
	if err == nil {
		packet := InitPacket(SC_UPDATE_VOTE, b.Bytes())
		tcpConnection.SendPacket(&packet)
	}

	log.Println("new connection")
}
