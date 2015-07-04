package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"gopkg.in/redis.v3"
	"strconv"
	"time"
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

func ConnectToRedis() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
}

func NewVote(name string, owner string) *Vote {
	id := strconv.FormatInt(time.Now().UnixNano(), 10)

	vote := &Vote{
		Name:  name,
		date:  time.Now().UnixNano(),
		owner: owner,
		Id:    id,
	}

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// retain readability with json
	serialized, err := json.Marshal(vote)

	if err == nil {
		fmt.Println("serialized data: ", string(serialized))

		err := client.Set(id, string(serialized), 0).Err()
		if err != nil {
			panic(err)
		}
	}

	return vote
}

func SaveUser(user auth.User) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// retain readability with json
	serialized, err := json.Marshal(user)

	if err == nil {
		fmt.Println("serialized data: ", string(serialized))

		err := client.Set("user:"+user.Id, string(serialized), 0).Err()
		if err != nil {
			panic(err)
		}
	}
}

func SaveAuthToken(at auth.AuthToken) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// retain readability with json
	serialized, err := json.Marshal(at)

	if err == nil {
		fmt.Println("serialized data: ", string(serialized))

		err := client.Set(at.HMAC, string(serialized), 0).Err()
		if err != nil {
			panic(err)
		}
	}
}

func GetAllUsers() []auth.User {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	users_keys, _ := client.Keys("user:*").Result()
	var users []auth.User
	for _, value := range users_keys {
		item, err := LoadUser(value)
		if err != nil {
			users = append(users, *item)
		}
	}
	return users
}

func LoadUser(id string) (*auth.User, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	data, err := client.Get(id).Result()
	if err != nil {
		return nil, errors.New("Not exist")
	} else {
		user := &auth.User{}
		json.Unmarshal([]byte(data), &user)
		return user, nil
	}
}
