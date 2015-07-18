package tcpsocket

import (
	"bytes"
	"encoding/json"
	"github.com/walkline/MoC-pulse-backend/events"
	"log"
	"strconv"
)

func ListenToEvents() {
	clients := make([]*events.SomeSocket, 0, 1024)
	clientsCounter := 0
	for {
		select {
		case newVoteEvent := <-*events.GetNewVoteChan():
			for _, client := range clients {
				client.NewVoteEvent <- &newVoteEvent
			}
		case voteUpdateEvent := <-*events.GetVoteUpdateChan():
			for _, client := range clients {
				client.VoteUpdEvent <- &voteUpdateEvent
			}
		case newSocketEvent := <-*events.GetNewSocketsChan():
			clientsCounter++
			newSocketEvent.Socket.Id = strconv.Itoa(clientsCounter)
			clients = append(clients, newSocketEvent.Socket)
		case socketClosedEvent := <-*events.GetClosedSocketsChan():
			id := socketClosedEvent.Socket.Id
			index := 0
			for i, socket := range clients {
				if socket.Id == id {
					index = i
				}
			}

			clients = append(clients[:index], clients[index+1:]...)
		}
	}
}

func (s *TcpSocket) ListenToEvents() {
	for {
		socketClosed := false

		select {
		case newVoteEvent := <-s.SomeSocket.NewVoteEvent:
			s.handleNewVoteEvent(newVoteEvent)
		case voteUpdateEvent := <-s.SomeSocket.VoteUpdEvent:
			s.handleVoteUpdateEvent(voteUpdateEvent)
		case _ = <-s.SomeSocket.CloseEvent:
			close(s.SomeSocket.CloseEvent)
			close(s.SomeSocket.NewVoteEvent)
			close(s.SomeSocket.VoteUpdEvent)

			socketClosed = true
		}

		if socketClosed {
			log.Println("stop listening...")
			break
		}
	}
}

func (s *TcpSocket) handleNewVoteEvent(e *events.NewVoteEvent) {
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(*e.Vote)
	if err == nil {
		packet := InitPacket(SC_NEW_VOTE, b.Bytes())
		s.SendPacket(&packet)
	}
}

func (s *TcpSocket) handleVoteUpdateEvent(e *events.VoteUpdateEvent) {
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(*e.Vote)
	if err == nil {
		packet := InitPacket(SC_UPDATE_VOTE, b.Bytes())
		s.SendPacket(&packet)
	}
}
