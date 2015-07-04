package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"log"
)

var endpoint string = "http://192.168.4.121:3000/api/trusted/profiles.json?name=PulsePush&token=1-K_I6DY1QbNknp-MXHN4QDhTmD1BQCgyesCoHfZExzABKdwvKOenIcisq7UubPprAcnZBrpP4qmu-j-nzlH_F8A"

type UserInterface interface {
}

func SaveUser(user auth.User) {
	client := ConnectToRedis()

	// retain readability with json
	serialized, err := json.Marshal(user)

	if err == nil {
		fmt.Println("serialized data: ", string(serialized))

		err := client.Set("user:"+user.Id, string(serialized), 0).Err()
		if err != nil {
			log.Fatal("Failed to set user into redis: ", err)
		}
	}

	client.Close()
}

func SaveAuthToken(at auth.AuthToken) {
	client := ConnectToRedis()
	defer client.Close()

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

	users_keys, err := client.Keys("user:*").Result()
	if err != nil {
		fmt.Println(err)
	}
	var users []auth.User
	for _, value := range users_keys {
		fmt.Println(value)
		item, err := LoadUser(value)
		if err == nil {
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

func GetUsers() ([]auth.User, error) {
	_, body, errs := gorequest.New().Get(endpoint).End()
	if errs != nil {
		return nil, errors.New("Can't get users from Auth provider")
	} else {

		var loaded []auth.User
		json.Unmarshal([]byte(body), &loaded)

		exist_users := GetAllUsers()

		var user_keys map[string]auth.User = map[string]auth.User{}

		for _, value := range exist_users {
			user_keys[value.Id] = value
		}

		var users []auth.User

		for _, value := range loaded {
			existed, ok := user_keys[value.Id]
			fmt.Println(ok)
			if !ok {
				users = append(users, value)
			} else {
				users = append(users, existed)
			}
		}

		return users, nil
	}
}
