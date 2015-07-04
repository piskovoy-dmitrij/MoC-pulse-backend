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
	Owner string
	Date  int64
}

type VoteResult struct {
	Id    string
	Value int
	Vote  string
	Date  int64
	User  string
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
