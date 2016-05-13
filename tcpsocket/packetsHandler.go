package tcpsocket

import (
	"bytes"
	"encoding/json"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/events"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/log"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/storage"
)

func (s *TcpSocket) ProcessPacket(packet *PulsePacket) {
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

func (s *TcpSocket) handleNewVote(packet *PulsePacket) {
	funcPrefix := "New vote creation"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	var params CSCreateVoteRequest
	log.Debug.Printf("%s: unmarshaling params...\n", funcPrefix)
	err := json.Unmarshal(packet.content, &params)
	if err != nil {
		log.Error.Printf("%s: unmarshaling params failed: %s\n", funcPrefix, err.Error())
		return
	}

	log.Debug.Printf("%s: adding new vote to storage...\n", funcPrefix)
	vote, err1 := storage.NewVote(params.Name, s.user.Id)	
	if err1 != nil {
		log.Error.Printf("%s: adding vote '%s' to storage failed: %s\n", funcPrefix, params.Name, err.Error())
		return
	}

	go func() {
		log.Debug.Printf("%s: getting users from storage...\n", funcPrefix)
		users, _ := storage.GetUsers()
		log.Debug.Printf("%s: removing vote creator from notification list...\n", funcPrefix)
		for p, v := range users {
			if s.user.Id == v.Id {
				users = append(users[:p], users[p+1:]...)
				log.Debug.Printf("%s: vote creator has been found and succesfully removed from the list\n", funcPrefix)
				break
			}
		}
		log.Debug.Printf("%s: sending notifications to users...\n", funcPrefix)
		notificationSender.Send(users, *vote)
	}()

	log.Debug.Printf("%s: getting vote result status...\n", funcPrefix)
	res, err2 := storage.GetVoteResultStatus(*vote, s.user)
	if err2 != nil {
		log.Error.Printf("%s: getting vote result status failed: %s\n", funcPrefix, err.Error())
		return
	}

	log.Debug.Printf("%s: sending new vote event...\n", funcPrefix)
	*events.GetNewVoteChan() <- events.NewVoteEvent{res}

	log.Info.Printf("%s: new vote '%s' has been succesfully created!\n", funcPrefix, params.Name)
}

func (s *TcpSocket) handleGetVote(packet *PulsePacket) {
	funcPrefix := "Getting vote results"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	var params CSGetVoteRequest
	log.Debug.Printf("%s: unmarshaling params...\n", funcPrefix)
	err := json.Unmarshal(packet.content, &params)
	if err != nil {
		log.Error.Printf("%s: unmarshaling params failed: %s\n", funcPrefix, err.Error())
		return
	}

	log.Debug.Printf("%s: getting vote with id '%s' from storage...\n", funcPrefix, params.Id)
	vote, err1 := storage.GetVote(params.Id)
	if err1 != nil {
		log.Error.Printf("%s: getting vote with id '%s' from storage failed: %s\n", funcPrefix, params.Id, err.Error())
		return
	}

	log.Info.Printf("%s: vote was successfully found: [%+v]\n", funcPrefix, vote)

	log.Debug.Printf("%s: getting vote result status...\n", funcPrefix)
	res, err2 := storage.GetVoteResultStatus(*vote, s.user)
	if err2 != nil {
		log.Error.Printf("%s: getting vote result status failed: %s\n", funcPrefix, err.Error())
		return
	}

	var b bytes.Buffer
	err = json.NewEncoder(&b).Encode(res)
	if err != nil {
		log.Error.Printf("%s: encoding result failed: %s\n", funcPrefix, err.Error())
		return
	}
	replyPacket := InitPacket(SC_GET_VOTE_RESULT, b.Bytes())
	s.SendPacket(&replyPacket)	
}

func (s *TcpSocket) handleGetVotes(packet *PulsePacket) {
	funcPrefix := "Getting all votes with results"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	log.Debug.Printf("%s: getting all votes with results from storage...\n", funcPrefix)
	votes, err := storage.GetAllVotesWithResult(s.user)
	if err != nil {
		log.Error.Printf("%s: getting all votes with results from storage failed: %s\n", funcPrefix, err.Error())
		return
	}

	res := storage.VotesStatus{
		Votes: votes,
	}

	var b bytes.Buffer
	err = json.NewEncoder(&b).Encode(res)
	if err != nil {
		log.Error.Printf("%s: encoding result failed: %s\n", funcPrefix, err.Error())
		return
	}
	replyPacket := InitPacket(SC_GET_VOTES_RESULT, b.Bytes())
	s.SendPacket(&replyPacket)
}

func (s *TcpSocket) handleVoteFor(packet *PulsePacket) {
	funcPrefix := "Processing voting"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	var params CSVoteForRequest
	log.Debug.Printf("%s: unmarshaling params...\n", funcPrefix)
	err := json.Unmarshal(packet.content, &params)
	if err != nil {
		log.Error.Printf("%s: unmarshaling params failed: %s\n", funcPrefix, err.Error())
		return
	}

	log.Debug.Printf("%s: getting vote with id '%s' from storage...\n", funcPrefix, params.Id)
	vote, err := storage.GetVote(params.Id)
	if err != nil {
		log.Error.Printf("%s: getting vote with id '%s' from storage failed: %s\n", funcPrefix, params.Id, err.Error())
		return
	}

	if storage.IsVotedByUser(*vote, s.user) {
		log.Warning.Printf("%s: user has already voted!\n", funcPrefix)
		return
	}

	log.Debug.Printf("%s: modifying vote...\n", funcPrefix)
	storage.VoteProcessing(*vote, s.user, params.ColorId)

	log.Debug.Printf("%s: getting vote result status...\n", funcPrefix)
	res, err1 := storage.GetVoteResultStatus(*vote, s.user)
	if err1 != nil {
		log.Error.Printf("%s: getting vote result status failed: %s\n", funcPrefix, err.Error())
		return
	}

	log.Debug.Printf("%s: sending vote update event...\n", funcPrefix)
	*events.GetVoteUpdateChan() <- events.VoteUpdateEvent{res}

	log.Info.Printf("%s: vote '%s' has been succesfully updated!\n", funcPrefix, vote.Name)
}

func (s *TcpSocket) handleAuth(packet *PulsePacket) {
	funcPrefix := "Authenticating user"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	var params CSAuthRequest
	log.Debug.Printf("%s: unmarshaling params...\n", funcPrefix)
	err := json.Unmarshal(packet.content, &params)
	if err != nil {
		log.Error.Printf("%s: unmarshaling params failed: %s\n", funcPrefix, err.Error())
		return
	}

	log.Debug.Printf("%s: authenticating user...\n", funcPrefix)
	user, authErr := storage.Authenticate(params.Token)
	if authErr != nil {
		log.Error.Printf("%s: user authentication failed\n", funcPrefix)
		return
	}
	s.user = *user

	var b bytes.Buffer
	err = json.NewEncoder(&b).Encode(s.user)
	if err == nil {
		log.Error.Printf("%s: encoding result failed: %s\n", funcPrefix, err.Error())
		return
	}	
	replyPacket := InitPacket(SC_AUTH, b.Bytes())
	s.SendPacket(&replyPacket)
}
