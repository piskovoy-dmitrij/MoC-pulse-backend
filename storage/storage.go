package storage

import (
	"gopkg.in/redis.v3"
	"encoding/json"
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
	
	defer client.Close()
	
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
	
	// retain readability with json
    serialized, err := json.Marshal(vote)

    if err == nil {
        fmt.Println("serialized data: ", string(serialized))
	}
	
	setKey(id, serialized)
	
	return &Vote{
		Name:  name,
		date:  time.Now().UnixNano(),
		owner: owner,
		Id:    id,
	}
}

func GetVoteByID(Id string) *Vote {
	vote, err := getKey(Id)
} 

func setKey(key string, value string) {
	client := ConnectToRedis()
	
	err := client.Set("key", "value", 0).Err()
    if err != nil {
        panic(err)
    }
}

func getKey(key string) {
	client := ConnectToRedis()
	
	val, err := client.Get(key).Result()
    if err == redis.Nil {
        fmt.Println(key + " does not exists")
    } else if err != nil {
        panic(err)
    } else {
        fmt.Println("returning by key: ", val)
    }
	
	return val
}