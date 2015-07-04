package storage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"log"
	"strconv"
	"time"
)

func NewVote(name string, owner string) *Vote {
	id := strconv.FormatInt(time.Now().UnixNano(), 10)

	vote := &Vote{
		Name:  name,
		Date:  time.Now().UnixNano(),
		Owner: owner,
		Id:    id,
	}

	client := ConnectToRedis()

	// retain readability with json
	serialized, err := json.Marshal(vote)

	if err == nil {
		fmt.Println("serialized data: ", string(serialized))

		err := client.Set("vote:"+id, string(serialized), 0).Err()
		if err != nil {
			panic(err)
		}
	}

	client.Close()

	return vote
}

func GetVote(id string) (*Vote, error) {
	client := ConnectToRedis()

	val, err := client.Get("vote:" + id).Result()

	client.Close()

	if err == nil {
		return nil, errors.New("Not found")
		fmt.Println("key " + id + " does not exists")
	} else if err != nil {
		log.Fatal("Failed to get vote by key "+id+": ", err)
		return nil, errors.New("Not found")
	}

	var vote Vote
	jsonString, err := base64.StdEncoding.DecodeString(val)
	err = json.Unmarshal(jsonString, &vote)
	if err != nil {
		log.Fatal("Failed to decode Vote: ", err)
	}	
	
	return &vote, nil
}

func SaveResult(result *VoteResult) {
	client := ConnectToRedis()
	defer client.Close()

	// retain readability with json
	serialized, err := json.Marshal(result)

	if err == nil {
		fmt.Println("serialized data: ", string(serialized))

		err := client.Set("result:"+result.Id, string(serialized), 0).Err()
		if err != nil {
			panic(err)
		}
	}
}

func NewResult(vote Vote, user auth.User, value int) *VoteResult {
	return &VoteResult{
		Id:    vote.Id + ":" + user.Id,
		Value: value,
		Vote:  vote.Id,
		Date:  time.Now().UnixNano(),
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

//func GetVoteResults(vote Vote) {
//	client := ConnectToRedis()

//	results_keys, err := client.Keys("result:" + vote.Id).Result()
//	if err != nil {
//		fmt.Println(err)
//	}
//	var results []VoteResult
//	for _, value := range results_keys {
//		fmt.Println(value)
//		item, err := LoadUser(value)
//		if err == nil {
//			users = append(users, *item)
//		}
//	}
//	client.Close()

//	return users
//}