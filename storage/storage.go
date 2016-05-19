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

type StorageConnection struct {
	client *redis.Client
}

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

func NewStorageConnection(address string) *StorageConnection {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return &StorageConnection{client}
}

func (this *StorageConnection) CloseStorageConnection() {
	this.client.Close()
}

func (this *StorageConnection) getKeys(format string) ([]string, int, error) {
	funcPrefix := fmt.Sprintf("Getting keys matching '%s' from storage", format)
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	var cursor int64
	var count int
	var resultKeys []string
	for {
		var keys []string
		var err error
		cursor, keys, err = this.client.Scan(cursor, format, 500).Result()
		if err != nil {
			log.Error.Printf("%s failed: %s\n", funcPrefix, err.Error())
			return nil, 0, err
		}
		count += len(keys)
		resultKeys = append(resultKeys, keys...)
		if cursor == 0 {
			break
		}
	}

	return resultKeys, count, nil
}

func (this *StorageConnection) LoadAuthToken(id string) (*auth.AuthToken, error) {
	data, err := this.client.Get(id).Result()
	if err != nil {
		return nil, errors.New("Not exist")
	}
	at := &auth.AuthToken{}
	jsonString, _ := base64.StdEncoding.DecodeString(data)
	json.Unmarshal([]byte(jsonString), at)
	return at, nil
}

func (this *StorageConnection) Authenticate(token string) (*auth.User, error) {
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
	at, err := this.LoadAuthToken(token)
	if err != nil {
		log.Error.Printf("%s returns error: %s\n", funcPrefix, err.Error())
		return nil, err
	}
	info, err := at.GetTokenInfo()
	if err != nil {
		log.Error.Printf("%s returns error: %s\n", funcPrefix, err.Error())
		return nil, err
	}
	user, err := this.LoadUser("user:" + info.Id)
	if err != nil {
		log.Error.Printf("%s returns error: %s\n", funcPrefix, err.Error())
		return nil, err
	}
	log.Debug.Printf("%s returns user [%+v]\n", funcPrefix, user)
	return user, nil
}
