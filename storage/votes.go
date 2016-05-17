package storage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/log"
)

func NewVote(name string, owner string) (*Vote, error) {
	funcPrefix := "Inserting new vote to storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	client := ConnectToRedis()
	defer client.Close()

	id := strconv.FormatInt(time.Now().UnixNano(), 10)

	vote := &Vote{
		Name:  name,
		Date:  time.Now().Unix(),
		Owner: owner,
		Id:    id,
		Voted: false,
	}

	log.Debug.Printf("%s: marshaling vote...\n", funcPrefix)
	serialized, err := json.Marshal(vote)
	if err != nil {
		log.Error.Printf("%s: marshaling vote failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}
	err = client.Set("vote:"+id, base64.StdEncoding.EncodeToString(serialized), 0).Err()
	if err != nil {
		log.Error.Printf("%s: inserting vote failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	return vote, nil
}

func GetVote(id string) (*Vote, error) {
	return LoadVote("vote:" + id)
}

func LoadVote(id string) (*Vote, error) {
	funcPrefix := "Getting vote from storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	client := ConnectToRedis()
	defer client.Close()

	log.Debug.Printf("%s: getting vote...\n", funcPrefix)
	val, err := client.Get(id).Result()
	if err != nil {
		log.Error.Printf("%s: getting vote failed: %s\n", funcPrefix, err.Error())
		return nil, errors.New("Not found")
	}

	var vote Vote
	jsonString, _ := base64.StdEncoding.DecodeString(val)
	log.Debug.Printf("%s: unmarshaling vote...\n", funcPrefix)
	err = json.Unmarshal(jsonString, &vote)
	if err != nil {
		log.Error.Printf("%s: unmarshaling vote failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	return &vote, nil
}

func GetAllVotesWithResult(user auth.User) ([]VoteWithResult, error) {
	funcPrefix := "Getting all votes from storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	client := ConnectToRedis()
	defer client.Close()

	log.Debug.Printf("%s: getting vote keys...\n", funcPrefix)
	votes_keys, err := client.Keys("vote:*").Result()
	if err != nil {
		log.Error.Printf("%s: getting vote keys failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	var votes []VoteWithResult

	log.Debug.Printf("%s: getting results of each vote by key...\n", funcPrefix)
	for _, value := range votes_keys {
		vote, error := LoadVote(value)
		if error == nil {
			item, error2 := GetVoteResultStatus(*vote, user)
			if error2 == nil {
				votes = append(votes, item.Vote)
			}
		}
	}

	return votes, nil
}

func VoteProcessing(vote Vote, user auth.User, value int) error {
	voteResult := NewResult(vote, user, value)

	return SaveResult(voteResult)
}

func SaveResult(result *VoteResult) error {
	funcPrefix := "Saving result to storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	client := ConnectToRedis()
	defer client.Close()

	log.Debug.Printf("%s: marshaling result...\n", funcPrefix)
	serialized, err := json.Marshal(result)

	if err != nil {
		log.Error.Printf("%s: marshaling result failed: %s\n", funcPrefix, err.Error())
		return err
	}
	err = client.Set("result:"+result.Id, base64.StdEncoding.EncodeToString(serialized), 0).Err()
	if err != nil {
		log.Error.Printf("%s: inserting result failed: %s\n", funcPrefix, err.Error())
		return err
	}

	return nil
}

func NewResult(vote Vote, user auth.User, value int) *VoteResult {
	return &VoteResult{
		Id:    vote.Id + ":" + user.Id,
		Value: value,
		Vote:  vote.Id,
		Date:  time.Now().UnixNano(),
	}
}

func LoadVoteResult(id string) (*VoteResult, error) {
	funcPrefix := "Getting result from storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	client := ConnectToRedis()
	defer client.Close()

	log.Debug.Printf("%s: getting result...\n", funcPrefix)
	data, err := client.Get(id).Result()
	if err != nil {
		log.Error.Printf("%s: getting result failed: %s\n", funcPrefix, err.Error())
		return nil, errors.New("Not exist")
	}

	jsonString, _ := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, errors.New("Not exist")
	}
	log.Debug.Printf("%s: unmarshaling vote result...\n", funcPrefix)
	voteResult := &VoteResult{}
	err = json.Unmarshal(jsonString, &voteResult)
	if err != nil {
		log.Error.Printf("%s: unmarshaling vote result failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	return voteResult, nil
}

func IsVotedByUser(vote Vote, user auth.User) bool {
	funcPrefix := "Checking if user has voted"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	client := ConnectToRedis()
	defer client.Close()

	_, err := client.Get("result:" + vote.Id + ":" + user.Id).Result()
	if err != nil {
		log.Debug.Printf("%s: user '%s' has not voted in vote with id '%s' yet\n", funcPrefix, user.Id, vote.Id)
		return false
	} else {
		log.Debug.Printf("%s: user '%s' has already voted in vote with id '%s'\n", funcPrefix, user.Id, vote.Id)
		return true
	}
}

func GetVoteResultStatus(vote Vote, user auth.User) (*VoteResultStatus, error) {
	funcPrefix := "Getting vote result status from storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	client := ConnectToRedis()
	defer client.Close()

	log.Debug.Printf("%s: getting vote result keys...\n", funcPrefix)
	results_keys, err := client.Keys("result:" + vote.Id + ":*").Result()
	if err != nil {
		log.Error.Printf("%s: getting vote result keys failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	var yellow int
	var red int
	var green int

	log.Debug.Printf("%s: getting vote result by result key...\n", funcPrefix)
	for _, value := range results_keys {
		item, err := LoadVoteResult(value)

		if err == nil {
			if item.Value == 0 {
				red = red + 1
			} else if item.Value == 1 {
				yellow = yellow + 1
			} else {
				green = green + 1
			}
		}
	}

	log.Debug.Printf("%s: getting vote owner...\n", funcPrefix)
	ownerUser, error := LoadUser("user:" + vote.Owner)
	if error != nil {
		log.Error.Printf("%s: getting vote owner failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	return &VoteResultStatus{
		Vote: VoteWithResult{
			Name:  vote.Name,
			Id:    vote.Id,
			Owner: *ownerUser,
			Date:  vote.Date,
			Voted: IsVotedByUser(vote, user),
			Result: Result{
				Yellow:    yellow,
				Green:     green,
				Red:       red,
				AllUsers:  UsersCount(),
				VoteUsers: yellow + green + red,
			},
		},
	}, nil
}
