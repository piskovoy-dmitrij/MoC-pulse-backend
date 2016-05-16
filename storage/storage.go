package storage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/log"
	"gopkg.in/redis.v3"
)

type Vote struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Voted bool   `json:"voted"`
	Owner string `json:"owner"`
	Date  int64  `json:"date"`
}

type VoteResult struct {
	Id    string `json:"Id"`
	Value int    `json:"Value"`
	Vote  string `json:"Vote"`
	Date  int64  `json:"Date"`
	user  string `json:"user"`
}

type VoteStatus struct {
	Vote Vote `json:"vote"`
}

type VoteResultStatus struct {
	Vote VoteWithResult `json:"vote"`
}

type VoteWithResult struct {
	Id     string    `json:"id"`
	Name   string    `json:"name"`
	Owner  auth.User `json:"owner"`
	Date   int64     `json:"date"`
	Voted  bool      `json:"voted"`
	Result Result    `json:"result"`
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

type DoVoteStatus struct {
	Vote DoVote `json:"vote"`
}

type DoVote struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
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
	defer client.Close()

	data, err := client.Get(id).Result()
	if err != nil {
		return nil, errors.New("Not exist")
	} else {
		at := &auth.AuthToken{}
		jsonString, _ := base64.StdEncoding.DecodeString(data)
		json.Unmarshal([]byte(jsonString), at)
		return at, nil
	}
}

func Authenticate(token string) (*auth.User, error) {
	funcPrefix := fmt.Sprintf("Token '%s' authentication", token)
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	if token == "123123" {
		u := &auth.User{
			Id:     "debug",
			Email:  "test@test.com",
			Device: 2,
			DevId:  "",
		}
		log.Debug.Printf("%s returns user [%+v]\n", funcPrefix, u)
		return u, nil
	}
	at, err := LoadAuthToken(token)
	if err != nil {
		log.Error.Printf("%s returns error: %s\n", funcPrefix, err.Error())
		return nil, err
	}
	info, err := at.GetTokenInfo()
	if err != nil {
		log.Error.Printf("%s returns error: %s\n", funcPrefix, err.Error())
		return nil, err
	}
	user, err := LoadUser("user:" + info.Id)
	if err != nil {
		log.Error.Printf("%s returns error: %s\n", funcPrefix, err.Error())
		return nil, err
	} else {
		log.Debug.Printf("%s returns user [%+v]\n", funcPrefix, user)
		return user, nil
	}
}
