package events

import (
	"github.com/walkline/MoC-pulse-backend/storage"
)

type SomeSocket struct {
	Id           string
	NewVoteEvent chan *NewVoteEvent
	VoteUpdEvent chan *VoteUpdateEvent
	CloseEvent   chan *SocketClosedEvent
}

type NewVoteEvent struct {
	Vote *storage.VoteResultStatus
}

type VoteUpdateEvent struct {
	Vote *storage.VoteResultStatus
}

type NewSocketEvent struct {
	Socket *SomeSocket
}

type SocketClosedEvent struct {
	Socket *SomeSocket
}

var newVoteChan chan NewVoteEvent
var voteUpdateChan chan VoteUpdateEvent
var newSocketsChan chan NewSocketEvent
var closedSocketsChan chan SocketClosedEvent

func GetNewVoteChan() *(chan NewVoteEvent) {
	if newVoteChan == nil {
		newVoteChan = make(chan NewVoteEvent, 1024)
	}

	return &newVoteChan
}

func GetVoteUpdateChan() *(chan VoteUpdateEvent) {
	if voteUpdateChan == nil {
		voteUpdateChan = make(chan VoteUpdateEvent, 1024)
	}

	return &voteUpdateChan
}

func GetNewSocketsChan() *(chan NewSocketEvent) {
	if newSocketsChan == nil {
		newSocketsChan = make(chan NewSocketEvent, 1024)
	}

	return &newSocketsChan
}

func GetClosedSocketsChan() *(chan SocketClosedEvent) {
	if closedSocketsChan == nil {
		closedSocketsChan = make(chan SocketClosedEvent, 1024)
	}

	return &closedSocketsChan
}
