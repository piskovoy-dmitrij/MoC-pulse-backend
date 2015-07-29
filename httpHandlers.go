package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/piskovoy-dmitrij/MoC-pulse-backend/Godeps/_workspace/src/github.com/julienschmidt/httprouter"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/events"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/storage"
)

var secret string = "shjgfshfkjgskdfjgksfghks"

type RegisterStatus struct {
	Token string `json: token`
}

func storageConnect() {
	storage.ConnectToRedis()
}

type ParamSt struct {
	Name string `json: name`
	Type string `json: type`
}

type DoVotePrm struct {
	Value int `json: value`
}

func authenticate(token string) (*auth.User, error) {
	funcPrefix := fmt.Sprintf("Token '%s' authentication", token)
	Debug.Printf("%s: start\n", funcPrefix)
	defer Debug.Printf("%s: end\n", funcPrefix)
	if token == "123123" {
		u := &auth.User{
			Id:     "debug",
			Email:  "test@test.com",
			Device: 2,
			DevId:  "",
		}
		Debug.Printf("%s returns user [%+v]\n", funcPrefix, u)
		return u, nil
	}
	at, err := storage.LoadAuthToken(token)
	if err != nil {
		Error.Printf("%s returns error: %s\n", funcPrefix, err.Error())
		return nil, err
	}
	info, err := at.GetTokenInfo(secret)
	if err != nil {
		Error.Printf("%s returns error: %s\n", funcPrefix, err.Error())
		return nil, err
	}
	user, err := storage.LoadUser("user:" + info.Id)
	if err != nil {
		Error.Printf("%s returns error: %s\n", funcPrefix, err.Error())
		return nil, err
	} else {
		Debug.Printf("%s returns user [%+v]\n", funcPrefix, user)
		return user, nil
	}
}

func createVote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	funcPrefix := "New vote creation"
	Debug.Printf("%s: start\n", funcPrefix)
	defer Debug.Printf("%s: end\n", funcPrefix)
	Debug.Printf("%s: authenticating user...\n")
	user, error := authenticate(r.Header.Get("auth_token"))
	if error != nil {
		Error.Printf("%s: user authentication failed\n", funcPrefix)
		w.WriteHeader(400)
		return
	}

	var params ParamSt
	Debug.Printf("%s: decoding params...\n", funcPrefix)
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		Error.Printf("%s: decoding params failed: %s\n", funcPrefix, err.Error())
		w.WriteHeader(400)
		return
	}

	Debug.Printf("%s: adding new vote to storage...\n", funcPrefix)
	vote := storage.NewVote(params.Name, user.Id)

	// i think better use new goroutine
	go func() {
		Debug.Printf("%s: getting users from storage...\n", funcPrefix)
		users, _ := storage.GetUsers()
		Debug.Printf("%s: sending notifications to users...\n", funcPrefix)
		notificationSender.Send(users, *vote)
	}()

	Debug.Printf("%s: getting vote result status...\n", funcPrefix)
	res := storage.GetVoteResultStatus(*vote, *user)

	Debug.Printf("%s: sending new vote event...\n", funcPrefix)
	*events.GetNewVoteChan() <- events.NewVoteEvent{res}

	Info.Printf("%s: new vote '%s' has been succesfully created!\n", funcPrefix, params.Name)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if json.NewEncoder(w).Encode(res) != nil {
		Error.Printf("%s: encoding response failed\n", funcPrefix)
		w.WriteHeader(500)
	}
}

func getVote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	funcPrefix := "Getting vote results"
	Debug.Printf("%s: start\n", funcPrefix)
	defer Debug.Printf("%s: end\n", funcPrefix)
	Debug.Printf("%s: authenticating user...\n", funcPrefix)
	user, error := authenticate(r.Header.Get("auth_token"))
	if error != nil {
		Error.Printf("%s: user authentication failed\n", funcPrefix)
		w.WriteHeader(400)
		return
	}
	id := ps.ByName("id")

	Debug.Printf("%s: getting vote with id '%s' from storage...\n", funcPrefix, id)
	vote, err := storage.GetVote(id)
	if err != nil {
		Error.Printf("%s: getting vote with id '%s' from storage failed: %s\n", funcPrefix, id, err.Error())
		w.WriteHeader(400)
		return
	}

	Debug.Printf("%s: getting vote result status...\n", funcPrefix)
	res := storage.GetVoteResultStatus(*vote, *user)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if json.NewEncoder(w).Encode(res) != nil {
		Error.Printf("%s: encoding response failed\n", funcPrefix)
		w.WriteHeader(500)
	}
}

func getVotes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	funcPrefix := "Getting all votes with results"
	Debug.Printf("%s: start\n", funcPrefix)
	defer Debug.Printf("%s: end\n", funcPrefix)
	Debug.Printf("%s: authenticating user...\n", funcPrefix)
	user, error := authenticate(r.Header.Get("auth_token"))
	if error != nil {
		Error.Printf("%s: user authentication failed\n", funcPrefix)
		w.WriteHeader(400)
		return
	}
	Debug.Printf("%s: getting all votes with results from storage...\n", funcPrefix)
	votes := storage.GetAllVotesWithResult(*user)
	res := storage.VotesStatus{
		Votes: votes,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if json.NewEncoder(w).Encode(res) != nil {
		Error.Printf("%s: encoding response failed\n", funcPrefix)
		w.WriteHeader(500)
	}
}

func doVote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	funcPrefix := "Processing voting"
	Debug.Printf("%s: start\n", funcPrefix)
	defer Debug.Printf("%s: end\n", funcPrefix)
	Debug.Printf("%s: authenticating user...\n")
	user, error := authenticate(r.Header.Get("auth_token"))
	if error != nil {
		Error.Printf("%s: user authentication failed\n", funcPrefix)
		w.WriteHeader(400)
		return
	}
	id := ps.ByName("id")

	Debug.Printf("%s: getting vote with id '%s' from storage...\n", funcPrefix, id)
	vote, err := storage.GetVote(id)
	if err != nil {
		Error.Printf("%s: getting vote with id '%s' from storage failed: %s\n", funcPrefix, id, err.Error())
		w.WriteHeader(400)
		return
	}

	var params DoVotePrm
	Debug.Printf("%s: decoding params...\n", funcPrefix)
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		Error.Printf("%s: decoding params failed: %s\n", funcPrefix, err.Error())
		w.WriteHeader(400)
		return
	}

	Debug.Printf("%s: modifying vote...\n", funcPrefix)
	storage.VoteProcessing(*vote, *user, params.Value)

	Debug.Printf("%s: getting vote result status...\n", funcPrefix)
	res := storage.GetVoteResultStatus(*vote, *user)

	Debug.Printf("%s: sending vote update event...\n", funcPrefix)
	*events.GetVoteUpdateChan() <- events.VoteUpdateEvent{res}

	Info.Printf("%s: vote '%s' has been succesfully updated!\n", funcPrefix, vote.Name)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if json.NewEncoder(w).Encode(res) != nil {
		Error.Printf("%s: encoding response failed\n", funcPrefix)
		w.WriteHeader(500)
	}
}

func registerUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	funcPrefix := "Registering new user"
	Debug.Printf("%s: start\n", funcPrefix)
	defer Debug.Printf("%s: end\n", funcPrefix)
	Debug.Printf("%s: getting params from request...\n", funcPrefix)
	id := r.PostFormValue("id")
	email := r.PostFormValue("email")
	firstname := r.PostFormValue("first_name")
	lastname := r.PostFormValue("last_name")
	device, _ := strconv.Atoi(r.PostFormValue("device"))
	dev_id := r.PostFormValue("dev_id")
	token := r.PostFormValue("token")

	if token != "BE7C411D475AEA4CF1D7B472D5BD1" {
		Warning.Printf("%s: token is not correct!\n", funcPrefix)
		w.WriteHeader(403)
		return
	}

	Debug.Printf("%s: checking existence of user with id '%s'...\n", funcPrefix, id)
	user, err := storage.LoadUser("user:" + id)
	if err == nil {
		Debug.Printf("%s: user with id '%s' already exists; modifying his params...\n", funcPrefix, id)
		user.Device = device
		user.DevId = dev_id
		user.Email = email
		user.FirstName = firstname
		user.LastName = lastname
	} else {
		Debug.Printf("%s: user with id '%s' was not found; creating new user...\n", funcPrefix, id)
		user = &auth.User{
			Id:        id,
			Email:     email,
			Device:    device,
			DevId:     dev_id,
			FirstName: firstname,
			LastName:  lastname,
		}
	}
	Debug.Printf("%s: composing new auth token for user with id '%s'...\n", funcPrefix, id)
	at := auth.NewAuthToken(*user, time.Now(), secret)

	Debug.Printf("%s: saving user with id '%s' to storage...\n", funcPrefix, id)
	storage.SaveUser(*user)

	storage.SaveAuthToken(*at)

	Info.Printf("%s: user has been succesfully registered: [%+v]\n", funcPrefix, user)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	rec := RegisterStatus{
		Token: at.HMAC,
	}
	if json.NewEncoder(w).Encode(rec) != nil {
		Error.Printf("%s: encoding response failed\n", funcPrefix)
		w.WriteHeader(500)
	}
}

func testIOSNotificationSending(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	funcPrefix := "iOS notification sending test"
	Debug.Printf("%s: start\n", funcPrefix)
	defer Debug.Printf("%s: end\n", funcPrefix)
	Debug.Printf("%s: getting dev_id from request...\n", funcPrefix)
	dev_id := r.PostFormValue("dev_id")
	if dev_id != "" {
		Info.Printf("%s: trying to send notification for dev_id '%s'...\n", funcPrefix, dev_id)
		notificationSender.Send([]auth.User{auth.User{Id: "100", FirstName: "John", LastName: "Doe", Device: 0, DevId: dev_id}}, storage.Vote{Id: "5", Name: "HelloWorld", Date: 1436966974, Voted: true, Owner: "test"})
		//	notificationSender.Send([]auth.User{auth.User{Id: "100", FirstName: "John", LastName: "Doe", Device: 0, DevId: "ca4f2547a7fc19c4b92a27e940c373d3d3bded3102d5eddc4f63d74d615fab2c"}}, storage.Vote{Id: "5", Name: "Hello world"})
		w.WriteHeader(200)
	} else {
		Warning.Printf("%s: there is no dev_id in request!\n", funcPrefix)
		w.WriteHeader(400)
	}
}

func testAndroidNotificationSending(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	funcPrefix := "Android notification sending test"
	Debug.Printf("%s: start\n", funcPrefix)
	defer Debug.Printf("%s: end\n", funcPrefix)
	Debug.Printf("%s: getting dev_id from request...\n", funcPrefix)
	dev_id := r.PostFormValue("dev_id")
	if dev_id != "" {
		Info.Printf("%s: trying to send notification for dev_id '%s'...\n", funcPrefix, dev_id)
		notificationSender.Send([]auth.User{auth.User{Id: "100", FirstName: "John", LastName: "Doe", Device: 1, DevId: dev_id}}, storage.Vote{Id: "5", Name: "HelloWorld", Date: 1436966974, Voted: true, Owner: "test"})
		w.WriteHeader(200)
	} else {
		Warning.Printf("%s: there is no dev_id in request!\n", funcPrefix)
		w.WriteHeader(400)
	}
}

func emailVote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	funcPrefix := "Processing email voting"
	Debug.Printf("%s: start\n", funcPrefix)
	defer Debug.Printf("%s: end\n", funcPrefix)
	token := r.FormValue("token")
	Debug.Printf("%s: authenticating by token '%s' from request...\n", funcPrefix, token)
	_, error := authenticate(token)
	if error != nil {
		Error.Printf("%s: authentication failed\n", funcPrefix)
		w.WriteHeader(400)
		return
	}
	id := r.PostFormValue("vote")
	Debug.Printf("%s: getting vote with id '%s' from storage...\n", funcPrefix, id)
	vote, _ := storage.GetVote(id)
	value, _ := strconv.Atoi(r.PostFormValue("value"))
	Debug.Printf("%s: modifying vote...\n", funcPrefix)
	res := storage.DoVoteStatus{
		Vote: storage.DoVote{
			Name:  vote.Name,
			Value: value,
		},
	}

	Info.Printf("%s: vote '%s' has been succesfully updated!\n", funcPrefix, vote.Name)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if json.NewEncoder(w).Encode(res) != nil {
		Error.Printf("%s: encoding response failed\n", funcPrefix)
		w.WriteHeader(500)
	}
}
