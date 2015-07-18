package tcpsocket

import (
	"github.com/walkline/MoC-pulse-backend/auth"
	"github.com/walkline/MoC-pulse-backend/events"
	"log"
	"net"
	"os"
)

type TcpSocket struct {
	events.SomeSocket

	user      auth.User
	conection *net.Conn
}

func (s *TcpSocket) SendPacket(p *PulsePucket) {
	(*s.conection).Write(p.ToSlice())
}

func ListenAndServer(host string) {
	l, err := net.Listen("tcp", host)
	if err != nil {
		log.Println("TcpSocket Error listening:", err.Error())
		os.Exit(1)
	}

	println("Starting tcp server...")

	defer l.Close()

	defer close(*events.GetNewVoteChan())
	defer close(*events.GetVoteUpdateChan())
	defer close(*events.GetNewSocketsChan())
	defer close(*events.GetClosedSocketsChan())

	go ListenToEvents()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("TcpSocket Error accepting: ", err.Error())
			os.Exit(1)
		}

		go HandleNewConnection(conn)
	}
}

func HandleNewConnection(c net.Conn) {
	s := TcpSocket{conection: &c}

	// closing in ListenToEvents()
	s.SomeSocket.NewVoteEvent = make(chan *events.NewVoteEvent)
	s.SomeSocket.VoteUpdEvent = make(chan *events.VoteUpdateEvent)
	s.SomeSocket.CloseEvent = make(chan *events.SocketClosedEvent)

	go s.ListenToEvents()

	*events.GetNewSocketsChan() <- events.NewSocketEvent{&s.SomeSocket}

	defer ConnectionClosed(&s)

	// read
	for {
		header := make([]byte, 6)
		headerLen, err := c.Read(header)
		if err != nil || headerLen != 6 {
			break
		}

		packet := InitPacketWithHeaderData(header)

		content := make([]byte, packet.size)
		contLen, err := c.Read(content)
		if err != nil || contLen != int(packet.size) {
			break
		}

		packet.content = content

		s.ProccesPacket(&packet)
	}
}

func ConnectionClosed(s *TcpSocket) {
	*events.GetClosedSocketsChan() <- events.SocketClosedEvent{&s.SomeSocket}
	s.SomeSocket.CloseEvent <- &events.SocketClosedEvent{&s.SomeSocket}
	log.Println("ConnectionClosed")
}
