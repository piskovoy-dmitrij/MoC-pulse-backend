package tcpsocket

import (
	"strconv"
	"bytes"
	"encoding/json"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/events"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/log"	
)

func ListenToEvents() {
	funcPrefix := "Processing socket managing events"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	clients := make([]*events.SomeSocket, 0, 1024)
	clientsCounter := 0
	for {
		select {
		case newVoteEvent := <-*events.GetNewVoteChan():
			log.Debug.Printf("%s: handling vote create event\n", funcPrefix)
			for _, client := range clients {
				client.NewVoteEvent <- &newVoteEvent
			}
		case voteUpdateEvent := <-*events.GetVoteUpdateChan():
			log.Debug.Printf("%s: handling vote update event\n", funcPrefix)
			for _, client := range clients {
				client.VoteUpdEvent <- &voteUpdateEvent
			}
		case newSocketEvent := <-*events.GetNewSocketsChan():
			log.Debug.Printf("%s: handling socket connection create event\n", funcPrefix)
			clientsCounter++
			newSocketEvent.Socket.Id = strconv.Itoa(clientsCounter)
			clients = append(clients, newSocketEvent.Socket)
		case socketClosedEvent := <-*events.GetClosedSocketsChan():
			log.Debug.Printf("%s: handling socket connection close event\n", funcPrefix)
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
	funcPrefix := "Processing events from socket connection"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	log.Debug.Printf("%s: starting listening...\n", funcPrefix)
	for {
		select {
		case newVoteEvent := <-s.SomeSocket.NewVoteEvent:
			log.Debug.Printf("%s: handling vote create event\n", funcPrefix)
			s.handleNewVoteEvent(newVoteEvent)
		case voteUpdateEvent := <-s.SomeSocket.VoteUpdEvent:
			log.Debug.Printf("%s: handling vote update event\n", funcPrefix)
			s.handleVoteUpdateEvent(voteUpdateEvent)
		case _ = <-s.SomeSocket.CloseEvent:
			log.Debug.Printf("%s: handling close event\n", funcPrefix)
			close(s.SomeSocket.CloseEvent)
			close(s.SomeSocket.NewVoteEvent)
			close(s.SomeSocket.VoteUpdEvent)
			log.Debug.Printf("%s: stopping listening...\n", funcPrefix)
			return
		}
	}
}

func (s *TcpSocket) handleNewVoteEvent(e *events.NewVoteEvent) {
	funcPrefix := "Handling vote create event"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(*e.Vote)
	if err != nil {
		log.Error.Printf("%s: encoding failed: %s\n", funcPrefix, err.Error())
		return
	}
	packet := InitPacket(SC_NEW_VOTE, b.Bytes())
	s.SendPacket(&packet)
}

func (s *TcpSocket) handleVoteUpdateEvent(e *events.VoteUpdateEvent) {
	funcPrefix := "Handling vote update event"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(*e.Vote)
	if err != nil {
		log.Error.Printf("%s: encoding failed: %s\n", funcPrefix, err.Error())
		return
	}
	packet := InitPacket(SC_UPDATE_VOTE, b.Bytes())
	s.SendPacket(&packet)	
}
