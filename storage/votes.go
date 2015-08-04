package storage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"strconv"
	"time"
)

func NewVote(name string, owner string) *Vote {
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

	serialized, err := json.Marshal(vote)

	if err == nil {
		err := client.Set("vote:"+id, base64.StdEncoding.EncodeToString(serialized), 0).Err()
		if err != nil {
			panic(err)
		}
	}

	return vote
}

func GetVote(id string) (*Vote, error) {
	return LoadVote("vote:" + id)
}

func LoadVote(id string) (*Vote, error) {
	client := ConnectToRedis()
	defer client.Close()

	val, err := client.Get(id).Result()

	if err != nil {
		fmt.Println("Vote not found in redis")
		return nil, errors.New("Not found")
	}

	var vote Vote
	jsonString, err := base64.StdEncoding.DecodeString(val)
	err = json.Unmarshal(jsonString, &vote)
	if err != nil {
		fmt.Println("Failed to decode Vote")
		return nil, err
	}

	return &vote, nil
}

func GetAllVotesWithResult(user auth.User) []VoteWithResult {
	client := ConnectToRedis()
	defer client.Close()

	votes_keys, err := client.Keys("vote:*").Result()
	if err != nil {
		fmt.Println(err)
	}

	var votes []VoteWithResult

	for _, value := range votes_keys {
		vote, error := LoadVote(value)
		if error == nil {
			item := GetVoteResultStatus(*vote, user)
			votes = append(votes, item.Vote)
		}
	}

	return votes
}

func VoteProcessing(vote Vote, user auth.User, value int) error {
	voteResult := NewResult(vote, user, value)

	return SaveResult(voteResult)
}

func SaveResult(result *VoteResult) error {
	client := ConnectToRedis()
	defer client.Close()

	// retain readability with json
	serialized, err := json.Marshal(result)

	if err == nil {
		err := client.Set("result:"+result.Id, base64.StdEncoding.EncodeToString(serialized), 0).Err()
		if err != nil {
			return err
		}
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
	client := ConnectToRedis()
	defer client.Close()

	data, err := client.Get(id).Result()
	jsonString, _ := base64.StdEncoding.DecodeString(data)

	if err != nil {
		return nil, errors.New("Not exist")
	} else {
		voteResult := &VoteResult{}
		json.Unmarshal(jsonString, &voteResult)

		return voteResult, nil
	}
}

func isVotedByUser(vote Vote, user auth.User) bool {
	client := ConnectToRedis()
	defer client.Close()

	_, err := client.Get("result:" + vote.Id + ":" + user.Id).Result()
	if err == nil {
		return false
	} else {
		return true
	}
}

func GetVoteResultStatus(vote Vote, user auth.User) *VoteResultStatus {
	client := ConnectToRedis()
	defer client.Close()

	results_keys, err := client.Keys("result:" + vote.Id + ":*").Result()
	if err != nil {
		fmt.Println(err)
	}

	var yellow int
	var red int
	var green int

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

	ownerUser,error := LoadUser("user:" + vote.Owner)
	
	if(error != nil) {
		fmt.Println(error)
		ownerUser = &user	
	}

	return &VoteResultStatus{
		Vote: VoteWithResult{
			Name:  vote.Name,
			Id:    vote.Id,
			Owner: *ownerUser,
			Date:  vote.Date,
			Voted: isVotedByUser(vote, user),
			Result: Result{
				Yellow:    yellow,
				Green:     green,
				Red:       red,
				AllUsers:  UsersCount(),
				VoteUsers: yellow + green + red,
			},
		},
	}
}
