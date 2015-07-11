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
		date:  time.Now().UnixNano(),
		owner: owner,
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
	client := ConnectToRedis()
	defer client.Close()

	val, err := client.Get("vote:" + id).Result()

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

func GetAllVotesWithResult() []VoteWithResult {
	client := ConnectToRedis()
	defer client.Close()

	votes_keys, err := client.Keys("vote:*").Result()
	if err != nil {
	    fmt.Println(err)
	}

	var votes []VoteWithResult

	for _, value := range votes_keys {
	    fmt.Println(value)
		vote,_ := GetVote(value)
	    item := GetVoteResultStatus(*vote)
	    votes = append(votes, item.Vote)
	}

	return votes


	



//	votes = append(votes, VoteWithResult{
//			Name: "Vote 1",
//			Id:   "sgdsfgsdfgsdfg",
//			Result: Result{
//				Yellow:    10,
//				Green:     5,
//				Red:       3,
//				AllUsers:  20,
//				VoteUsers: 18,
//			},
//		})
	
//	votes = append(votes, VoteWithResult{
//			Name: "Vote 2",
//			Id:   "sgdsfgsdfgsdfg",
//			Result: Result{
//				Yellow:    10,
//				Green:     5,
//				Red:       3,
//				AllUsers:  20,
//				VoteUsers: 18,
//			},
//		})
	
//	return votes
}

func VoteProccessing(vote Vote, user auth.User, value int) *DoVoteStatus {
	voteResult := NewResult(vote, user, value)
	
	SaveResult(voteResult)
	
	return &DoVoteStatus{
		Vote: DoVote{
			Name:  vote.Name,
			Value: value,
		},
	}
	
}

func SaveResult(result *VoteResult) {
	client := ConnectToRedis()
	defer client.Close()

	// retain readability with json
	serialized, err := json.Marshal(result)

	if err == nil {
		err := client.Set("result:" + result.Id, base64.StdEncoding.EncodeToString(serialized), 0).Err()
		if err != nil {
			panic(err)
		}
	}
}

func NewResult(vote Vote, user auth.User, value int) *VoteResult {
	return &VoteResult{
		Id:    vote.Id + ":" + user.Id,
		value: value,
		vote:  vote.Id,
		date:  time.Now().UnixNano(),
	}
}

func LoadVoteResult(id string) (*VoteResult, error) {
	client := ConnectToRedis()
	defer client.Close()

	data, err := client.Get(id).Result()

	if err != nil {
		return nil, errors.New("Not exist")
	} else {
		voteResult := &VoteResult{}
		json.Unmarshal([]byte(data), &voteResult)
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

func GetVoteResultStatus(vote Vote) *VoteResultStatus {
	client := ConnectToRedis()
	defer client.Close()

	results_keys, err := client.Keys("result:" + vote.Id).Result()
	if err != nil {
		fmt.Println(err)
	}
	
	var yellow int
	var red int
	var green int
		
	for _, value := range results_keys {
		fmt.Println(value)
		item, err := LoadVoteResult(value)
		if err == nil {
			if(item.value == 0) {
				red = red + 1
			} else if(item.value == 1) {
				yellow = yellow + 1
			} else {
				green = green + 1
			}
		}
	}

	return &VoteResultStatus{
		Vote: VoteWithResult{
			Name: vote.Name,
			Id:   vote.Id,
			Result: Result{
				Yellow:    yellow,
				Green:     green,
				Red:       red,
				AllUsers:  20,
				VoteUsers: yellow + green + red,
			},
		},
	}
}