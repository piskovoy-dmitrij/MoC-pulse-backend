package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/parnurzeal/gorequest"
)

var endpoint string = "http://192.168.4.121:3000/api/trusted/profiles.json?name=PulsePush&token=1-K_I6DY1QbNknp-MXHN4QDhTmD1BQCgyesCoHfZExzABKdwvKOenIcisq7UubPprAcnZBrpP4qmu-j-nzlH_F8A"

type UserInterface interface {
}

func GetUsers() (map[string]*User, error) {
	//	response, err := http.Get("http://192.168.4.121:3000/api/trusted/profiles?name=PulsePush&token=1-K_I6DY1QbNknp-MXHN4QDhTmD1BQCgyesCoHfZExzABKdwvKOenIcisq7UubPprAcnZBrpP4qmu-j-nzlH_F8A")
	_, body, errs := gorequest.New().Get(endpoint).End()
	if errs != nil {
		return nil, errors.New("Can't get users from Auth provider")
	} else {

		var loaded []User
		json.Unmarshal([]byte(body), &loaded)

		for _, value := range loaded {
			fmt.Println(value.Email)
		}

		users := make(map[string]*User)
		return users, nil
	}
}
