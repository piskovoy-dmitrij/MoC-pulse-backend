package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
//	"strconv"
//	"time"
	"log"
)

func SaveUser(user auth.User) {
	client := ConnectToRedis()
	
	// retain readability with json
    serialized, err := json.Marshal(user)

    if err == nil {		
		err := client.Set("user:" + user.Id, string(serialized), 0).Err()
	    if err != nil {
	        log.Fatal("Failed to set user into redis: ", err)
	    }
	}
	
	client.Close()
} 

func SaveAuthToken(at auth.AuthToken) {
	client := ConnectToRedis()
	
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

//func GetAuthToken(token string) *auth.AuthToken {
//	client := ConnectToRedis()
	
	
//}

func GetAllUsers() []auth.User {
	client := ConnectToRedis()
	
	users_keys, _ := client.Keys("user:*").Result()
	var users []auth.User
	for _, value := range users_keys {
		item, err := LoadUser(value)
		if err != nil {
			users = append(users, *item)
		}
	}
	client.Close()

	return users
}

func LoadUser(id string) (*auth.User, error) {
	client := ConnectToRedis()
	
	data, err := client.Get(id).Result()
	client.Close()
	
	if err != nil {
		return nil, errors.New("Not exist")
	} else {
		user := &auth.User{}
		json.Unmarshal([]byte(data), &user)
		return user, nil
	}
}