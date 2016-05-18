package storage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/parnurzeal/gorequest"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/log"
)

var endpoint string = "https://id.masterofcode.com/api/trusted/profiles.json?name=PulsePush&token=1-K_I6DY1QbNknp-MXHN4QDhTmD1BQCgyesCoHfZExzABKdwvKOenIcisq7UubPprAcnZBrpP4qmu-j-nzlH_F8A"

func (this *StorageConnection) SaveUser(user auth.User) {
	funcPrefix := "Saving user to storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	log.Debug.Printf("%s: marshaling user...\n", funcPrefix)
	serialized, err := json.Marshal(user)
	if err != nil {
		log.Error.Printf("%s: marshaling user failed: %s\n", funcPrefix, err.Error())
		return
	}
	err = this.client.Set("user:"+user.Id, base64.StdEncoding.EncodeToString(serialized), 0).Err()
	if err != nil {
		log.Error.Printf("%s: inserting user failed: %s\n", funcPrefix, err.Error())
	}
}

func (this *StorageConnection) SaveAuthToken(at auth.AuthToken) {
	funcPrefix := "Saving auth token to storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	log.Debug.Printf("%s: marshaling auth token...\n", funcPrefix)
	serialized, err := json.Marshal(at)
	if err != nil {
		log.Error.Printf("%s: marshaling auth token failed: %s\n", funcPrefix, err.Error())
		return
	}
	err = this.client.Set(at.HMAC, base64.StdEncoding.EncodeToString(serialized), 0).Err()
	if err != nil {
		log.Error.Printf("%s: inserting auth token failed: %s\n", funcPrefix, err.Error())
	}
}

func (this *StorageConnection) GetAllUsers() ([]auth.User, error) {
	funcPrefix := "Getting all users from storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	log.Debug.Printf("%s: getting user keys...\n", funcPrefix)
	userKeys, _, err := this.getKeys("user:*")
	if err != nil {
		log.Error.Printf("%s: getting user keys failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	log.Debug.Printf("%s: getting users by keys...\n", funcPrefix)
	data, err := this.client.MGet(userKeys).Result()
	if err != nil {
		log.Error.Printf("%s: getting users failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	var users []auth.User
	log.Debug.Printf("%s: unmarshaling users...\n", funcPrefix)
	for _, value := range data {
		user := &auth.User{}
		err = json.Unmarshal([]byte(base64.StdEncoding.DecodeString(value)), user)
		if err != nil {
			log.Error.Printf("%s: unmarshaling user failed: %s\n", funcPrefix, err.Error())
			return nil, err
		}
		users = append(users, *user)
	}

	return users, nil
}

func (this *StorageConnection) LoadUser(id string) (*auth.User, error) {
	funcPrefix := "Loading user from storage"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	log.Debug.Printf("%s: getting user...\n", funcPrefix)
	data, err := this.client.Get(id).Result()
	if err != nil {
		log.Error.Printf("%s: getting user failed: %s\n", funcPrefix, err.Error())
		return nil, errors.New("Not exist")
	}

	user := &auth.User{}
	log.Debug.Printf("%s: unmarshaling user...\n", funcPrefix)
	err = json.Unmarshal([]byte(base64.StdEncoding.DecodeString(data)), user)
	if err != nil {
		log.Error.Printf("%s: unmarshaling user failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	return user, nil
}

func (this *StorageConnection) GetUsers() ([]auth.User, error) {
	funcPrefix := "Getting users with auth"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	log.Debug.Printf("%s: requesting users...\n", funcPrefix)
	_, body, errs := gorequest.New().Get(endpoint).End()
	if errs != nil {
		log.Error.Printf("%s: getting users from Auth provider failed\n", funcPrefix)
		return nil, errors.New("Can't get users from Auth provider")
	}

	var loadedUsers []auth.User
	log.Debug.Printf("%s: unmarshaling users...\n", funcPrefix)
	err := json.Unmarshal([]byte(body), &loadedUsers)
	if err != nil {
		log.Error.Printf("%s: unmarshaling users failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	existUsers, err := this.GetAllUsers()
	if err != nil {
		log.Error.Printf("%s failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	userKeys := map[string]auth.User{}
	for _, value := range existUsers {
		userKeys[value.Id] = value
	}

	var users []auth.User
	log.Debug.Printf("%s: populating result slice...\n", funcPrefix)
	for _, value := range loadedUsers {
		user, ok := userKeys[value.Id]
		if !ok {
			users = append(users, value)
		} else {
			users = append(users, user)
		}
	}

	return users, nil
}

func (this *StorageConnection) UsersCount() int {
	_, count, _ := this.getKeys("user:*")
	return count
}
