package tcpsocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/events"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/storage"
)

var secret string = "shjgfshfkjgskdfjgksfghks"

func authenticate(token string) (*auth.User, error) {
	if token == "123123" {
		u := &auth.User{
			Id:     "debug",
			Email:  "test@test.com",
			Device: 2,
			DevId:  "",
		}
		return u, nil
	}
	at, err := storage.LoadAuthToken(token)
	if err != nil {
		return nil, err
	}
	info, err := at.GetTokenInfo(secret)
	if err != nil {
		return nil, err
	}
	user, err := storage.LoadUser("user:" + info.Id)
	if err != nil {
		return nil, err
	} else {
		return user, nil
	}
}

func (s *TcpSocket) ProccesPacket(packet *PulsePucket) {
	switch packet.opcode {
	case CS_AUTH:
		s.handleAuth(packet)
	case CS_CREATE_VOTE:
		s.handleNewVote(packet)
	case CS_GET_VOTE:
		s.handleGetVote(packet)
	case CS_GET_VOTES:
		s.handleGetVotes(packet)
	case CS_VOTE_FOR:
		s.handleVoteFor(packet)
	}
}

func (s *TcpSocket) handleNewVote(packet *PulsePucket) {
	var params CSCreateVoteRequest
	err := json.Unmarshal(packet.content, &params)
	if err != nil {
		fmt.Println(err)
		return
	}

	vote := storage.NewVote(params.Name, s.user.Id)

	res := storage.GetVoteResultStatus(*vote, s.user)

	go func() {
		users, _ := storage.GetUsers()
		notificationSender.Send(users, *vote)
	}()

	*events.GetNewVoteChan() <- events.NewVoteEvent{res}
}

func (s *TcpSocket) handleGetVote(packet *PulsePucket) {
	var params CSGetVoteRequest
	err := json.Unmarshal(packet.content, &params)
	if err != nil {
		fmt.Println(err)
		return
	}

	vote, err := storage.GetVote(params.Id)

	if err != nil {
		fmt.Println(err)
		return
	}

	res := storage.GetVoteResultStatus(*vote, s.user)

	var b bytes.Buffer
	err = json.NewEncoder(&b).Encode(*res)
	if err == nil {
		packet := InitPacket(SC_GET_VOTE_RESULT, b.Bytes())
		s.SendPacket(&packet)
	}
}

func (s *TcpSocket) handleGetVotes(packet *PulsePucket) {
	votes := storage.GetAllVotesWithResult(s.user)
	res := storage.VotesStatus{
		Votes: votes,
	}

	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(res)
	if err == nil {
		packet := InitPacket(SC_GET_VOTES_RESULT, b.Bytes())
		s.SendPacket(&packet)
	}
}

func (s *TcpSocket) handleVoteFor(packet *PulsePucket) {
	var params CSVoteForRequest
	err := json.Unmarshal(packet.content, &params)
	if err != nil {
		fmt.Println(err)
		return
	}

	vote, err := storage.GetVote(params.Id)
	if err != nil {
		fmt.Println(err)
		return
	}

	storage.VoteProcessing(*vote, s.user, params.ColorId)

	res := storage.GetVoteResultStatus(*vote, s.user)

	*events.GetVoteUpdateChan() <- events.VoteUpdateEvent{res}
}

func (s *TcpSocket) handleAuth(packet *PulsePucket) {
	var params CSAuthRequest
	err := json.Unmarshal(packet.content, &params)
	if err != nil {
		fmt.Println(err)
		return
	}

	user, authErr := authenticate(params.Token)

	if authErr == nil {
		s.user = *user
	}

	var b bytes.Buffer
	err = json.NewEncoder(&b).Encode(s.user)
	if err == nil {
		packet := InitPacket(SC_AUTH, b.Bytes())
		s.SendPacket(&packet)
	}
}
