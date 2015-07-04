package auth

import (
	"errors"
	"net/http"
)

var endpoint string = "http://"

func getUsers() (map[string]*User, error) {
	_, err := http.Get(endpoint)
	if err != nil {
		return nil, errors.New("Can't get users from Auth provider")
	} else {
		//TODO parse response
		users := make(map[string]*User)
		return users, nil
	}
}
