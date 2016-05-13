package storage

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/parnurzeal/gorequest"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/log"
)

var endpoint string = "https://id.masterofcode.com/api/trusted/profiles.json?name=PulsePush&token=1-K_I6DY1QbNknp-MXHN4QDhTmD1BQCgyesCoHfZExzABKdwvKOenIcisq7UubPprAcnZBrpP4qmu-j-nzlH_F8A"

type UserInterface interface {
}

func SaveUser(user auth.User) {
	funcPrefix := "Saving user to storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	client := ConnectToRedis()
	defer client.Close()

	log.Debug.Printf("%s: marshaling user...\n", funcPrefix)
	serialized, err := json.Marshal(user)
	if err != nil {
		log.Error.Printf("%s: marshaling user failed: %s\n", funcPrefix, err.Error())
		return
	}
	err = client.Set("user:"+user.Id, base64.StdEncoding.EncodeToString(serialized), 0).Err()
	if err != nil {
		log.Error.Printf("%s: inserting user failed: %s\n", funcPrefix, err.Error())
	}
}

func SaveAuthToken(at auth.AuthToken) {
	funcPrefix := "Saving auth token to storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	client := ConnectToRedis()
	defer client.Close()

	log.Debug.Printf("%s: marshaling auth token...\n", funcPrefix)
	serialized, err := json.Marshal(at)
	if err != nil {
		log.Error.Printf("%s: marshaling auth token failed: %s\n", funcPrefix, err.Error())
		return
	}
	err = client.Set(at.HMAC, base64.StdEncoding.EncodeToString(serialized), 0).Err()
	if err != nil {
		log.Error.Printf("%s: inserting auth token failed: %s\n", funcPrefix, err.Error())
	}
}

func GetAllUsers() ([]auth.User, error) {
	funcPrefix := "Getting all users from storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	client := ConnectToRedis()
	defer client.Close()

	log.Debug.Printf("%s: getting user keys...\n", funcPrefix)
	users_keys, err := client.Keys("user:*").Result()
	if err != nil {
		log.Error.Printf("%s: getting user keys failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	var users []auth.User

	log.Debug.Printf("%s: getting each user by key...\n", funcPrefix)
	for _, value := range users_keys {
		item, err := LoadUser(value)
		if err == nil {
			users = append(users, *item)
		}
	}

	return users, nil
}

func LoadUser(id string) (*auth.User, error) {
	funcPrefix := "Getting user from storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	client := ConnectToRedis()
	defer client.Close()

	log.Debug.Printf("%s: getting user...\n", funcPrefix)
	data, err := client.Get(id).Result()
	if err != nil {
		log.Error.Printf("%s: getting user failed: %s\n", funcPrefix, err.Error())
		return nil, errors.New("Not exist")
	}

	user := &auth.User{}
	jsonString, _ := base64.StdEncoding.DecodeString(data)
	log.Debug.Printf("%s: unmarshaling user...\n", funcPrefix)
	err = json.Unmarshal([]byte(jsonString), user)
	if err != nil {
		log.Error.Printf("%s: unmarshaling user failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	return user, nil	
}

func GetUsers() ([]auth.User, error) {
	funcPrefix := "Getting users with auth"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	log.Debug.Printf("%s: requesting users...\n", funcPrefix)
	_, body, errs := gorequest.New().Get(endpoint).End()
	if errs != nil {
		log.Error.Printf("%s: getting users from Auth provider failed\n", funcPrefix)
		return nil, errors.New("Can't get users from Auth provider")
	}

	var loaded []auth.User
	log.Debug.Printf("%s: unmarshaling users...\n", funcPrefix)
	err := json.Unmarshal([]byte(body), &loaded)
	if err != nil {
		log.Error.Printf("%s: unmarshaling users failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	exist_users, err := GetAllUsers()

	var user_keys map[string]auth.User = map[string]auth.User{}

	for _, value := range exist_users {
		user_keys[value.Id] = value
	}
	var users []auth.User

	log.Debug.Printf("%s: populating result slice...\n", funcPrefix)
	for _, value := range loaded {
		existed, ok := user_keys[value.Id]
		if !ok {
			users = append(users, value)
		} else {
			users = append(users, existed)
		}
	}

	return users, nil
}

func UsersCount() int {
	users, _ := GetAllUsers()
	return cap(users)
}
