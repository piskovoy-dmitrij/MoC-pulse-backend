package tcpsocket

import (
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/events"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/notification"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/log"
	"net"
	"os"
)

var notificationSender *notification.Sender

type TcpSocket struct {
	events.SomeSocket

	user      auth.User
	conection *net.Conn
}

func (s *TcpSocket) SendPacket(p *PulsePacket) {
	(*s.conection).Write(p.ToSlice())
}

func ListenAndServer(host string, ns *notification.Sender) {
	notificationSender = ns
	l, err := net.Listen("tcp", host)
	if err != nil {
		log.Error.Printf("TcpSocket Error listening: %s\n", err.Error())
		os.Exit(1)
	}

	defer l.Close()

	defer close(*events.GetNewVoteChan())
	defer close(*events.GetVoteUpdateChan())
	defer close(*events.GetNewSocketsChan())
	defer close(*events.GetClosedSocketsChan())

	go ListenToEvents()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error.Printf("TcpSocket Error accepting: %s\n", err.Error())
			os.Exit(1)
		}

		go HandleNewConnection(conn)
	}
}

func HandleNewConnection(c net.Conn) {
	funcPrefix := "Handling new connection"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)
	s := TcpSocket{conection: &c}

	// closing in ListenToEvents()
	s.SomeSocket.NewVoteEvent = make(chan *events.NewVoteEvent)
	s.SomeSocket.VoteUpdEvent = make(chan *events.VoteUpdateEvent)
	s.SomeSocket.CloseEvent = make(chan *events.SocketClosedEvent)

	go s.ListenToEvents()

	log.Debug.Printf("%s: sending new socket connection event...\n", funcPrefix)
	*events.GetNewSocketsChan() <- events.NewSocketEvent{&s.SomeSocket}

	defer ConnectionClosed(&s)
	defer log.Debug.Printf("%s: closing connection...\n", funcPrefix)

	// read
	log.Debug.Printf("%s: start reading packets from socket connection\n", funcPrefix)
	for {
		header := make([]byte, 6)
		headerLen, err := c.Read(header)
		if err != nil || headerLen != 6 {
			break
		}

		packet := InitPacketWithHeaderData(header)
		if packet.size > 0 {
			content := make([]byte, packet.size)
			contLen, err := c.Read(content)
			if err != nil || contLen != int(packet.size) {
				break
			}
			packet.content = content
		}

		s.ProcessPacket(&packet)
	}
	log.Debug.Printf("%s: finish reading packets from socket connection\n", funcPrefix)	
}

func ConnectionClosed(s *TcpSocket) {
	*events.GetClosedSocketsChan() <- events.SocketClosedEvent{&s.SomeSocket}
	s.SomeSocket.CloseEvent <- &events.SocketClosedEvent{&s.SomeSocket}
}
