package storage

import (
	"encoding/json"
	"encoding/base64"
	"errors"
	"fmt"
//	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"strconv"
	"time"
	"log"
)

func NewVote(name string, owner string) *Vote {
	id := strconv.FormatInt(time.Now().UnixNano(), 10)

	vote := &Vote{
		Name:  name,
		date:  time.Now().UnixNano(),
		owner: owner,
		Id:    id,
	}

	client := ConnectToRedis()
	
	// retain readability with json
	serialized, err := json.Marshal(vote)

	if err == nil {
		fmt.Println("serialized data: ", string(serialized))

		err := client.Set("vote:" + id, string(serialized), 0).Err()
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
        fmt.Println("key "+ id + " does not exists")
    } else if err != nil {
		log.Fatal("Failed to get vote by key " + id + ": ", err)
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

func GetVoteWithResults(id string) (*Vote, error) {
	vote := Vote {
		Id:   id,
		Name: "debug",
	}
	
	return &vote, nil
}