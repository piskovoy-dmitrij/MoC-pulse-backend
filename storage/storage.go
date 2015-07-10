package storage

import (
	"encoding/json"
	"errors"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"gopkg.in/redis.v3"
)

type Vote struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	owner string
	date  int64
}

type VoteResult struct {
	Id    string
	value int
	vote  string
	date  int64
	user  string
}

type VoteStatus struct {
	Vote Vote `json:"vote"`
}

type VoteResultStatus struct {
	Vote VoteWithResult `json:"vote"`
}

type VoteWithResult struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Result Result `json:"result"`
}

type Result struct {
	Yellow    int `json:"yellow"`
	Green     int `json:"green"`
	Red       int `json:"red"`
	AllUsers  int `json:"all_users"`
	VoteUsers int `json:"vote_users"`
}

type VotesStatus struct {
	Votes []VoteWithResult `json:"votes"`
}

func ConnectToRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	
	return client
}

func LoadAuthToken(id string) (*auth.AuthToken, error) {
	client := ConnectToRedis()
	
	data, err := client.Get(id).Result()
	if err != nil {
		return nil, errors.New("Not exist")
	} else {
		at := &auth.AuthToken{}
		json.Unmarshal([]byte(data), &at)
		return at, nil
	}
}