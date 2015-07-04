package storage

import (
	"encoding/json"
	"fmt"
	"errors"
	"gopkg.in/redis.v3"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
//	"strconv"
//	"time"
)

type Vote struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	owner string
	date  int64
}

type VoteResult struct {
	id    string
	value int
	vote  string
	date  int64
	user  string
}

func ConnectToRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

//	pong, err := client.Ping().Result()
//	fmt.Println(pong, err)
	
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