package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
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

func SaveVote() {
	
}