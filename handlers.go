package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/storage"
	"net/http"
	"strconv"
	"time"
)

var secret string = "shjgfshfkjgskdfjgksfghks"

type RegisterStatus struct {
	Token string `json: token`
}

type VoteStatus struct {
	Vote storage.Vote `json:"vote"`
}

type VoteResultStatus struct {
	Vote VoteWithResult `json:"vote"`
}

type VoteWithResult struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Result Result `json:"result"`
}

type Result struct {
	Yellow    int `json:"yellow"`
	Green     int `json:"green"`
	Red       int `json:"red"`
	AllUsers  int `json:"all_users"`
	VoteUsers int `json:"vote_users"`
}

type VotesStatus struct {
	Votes []VoteWithResult `json:"votes"`
}

type DoVoteStatus struct {
	Vote DoVote `json:"vote"`
}

type DoVote struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func storageConnect() {
	storage.ConnectToRedis()
}

func authenticate(token string) (*auth.User, error) {
	if token == "123123" {
		return &auth.User{
			Id:     "debug",
			Email:  "test@test.com",
			Device: 2,
			DevId:  "",
		}, nil
	}
	//TODO load AuthToken from redis by token
	at := &auth.AuthToken{
		Info: token,
		HMAC: "ggggg",
	}
	info, err := at.GetTokenInfo(secret)
	if err != nil {
		return nil, err
	}
	//TODO load user
	return &auth.User{
		Id:     info.Id,
		Email:  "test@test.com",
		Device: 2,
		DevId:  "",
	}, nil
}

func createVote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, error := authenticate(r.Header.Get("auth_token"))
	if error != nil {
		w.WriteHeader(400)
		return
	}
	name := r.PostFormValue("name")
	vote := storage.NewVote(name, user.Id)
	res := VoteStatus{
		Vote: *vote,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if json.NewEncoder(w).Encode(res) != nil {
		w.WriteHeader(500)
	}
}

func getVote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, error := authenticate(r.Header.Get("auth_token"))
	if error != nil {
		w.WriteHeader(400)
		return
	}
	id := ps.ByName("id")
	vote := storage.Vote{
		Id:   id,
		Name: "debug",
	}
	res := VoteResultStatus{
		Vote: VoteWithResult{
			Name: vote.Name,
			Id:   vote.Id,
			Result: Result{
				Yellow:    10,
				Green:     5,
				Red:       3,
				AllUsers:  20,
				VoteUsers: 18,
			},
		},
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if json.NewEncoder(w).Encode(res) != nil {
		w.WriteHeader(500)
	}
}

func getVotes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, error := authenticate(r.Header.Get("auth_token"))
	if error != nil {
		w.WriteHeader(400)
		return
	}
	votes := [...]VoteWithResult{
		VoteWithResult{
			Name: "Vote 1",
			Id:   "sgdsfgsdfgsdfg",
			Result: Result{
				Yellow:    10,
				Green:     5,
				Red:       3,
				AllUsers:  20,
				VoteUsers: 18,
			},
		},
		VoteWithResult{
			Name: "Vote 2",
			Id:   "sgdssdfgggsdfgsdfg",
			Result: Result{
				Yellow:    10,
				Green:     5,
				Red:       3,
				AllUsers:  20,
				VoteUsers: 18,
			},
		},
	}
	res := VotesStatus{
		Votes: votes[0:2],
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if json.NewEncoder(w).Encode(res) != nil {
		w.WriteHeader(500)
	}
}

func doVote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, error := authenticate(r.Header.Get("auth_token"))
	if error != nil {
		w.WriteHeader(400)
		return
	}
	id := ps.ByName("id")
	vote := storage.Vote{
		Id:   id,
		Name: "debug",
	}
	value, _ := strconv.Atoi(r.PostFormValue("value"))
	res := DoVoteStatus{
		Vote: DoVote{
			Name:  vote.Name,
			Value: value,
		},
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if json.NewEncoder(w).Encode(res) != nil {
		w.WriteHeader(500)
	}
}

func registerUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id := r.PostFormValue("id")
	email := r.PostFormValue("email")
	firstname := r.PostFormValue("first_name")
	lastname := r.PostFormValue("last_name")
	device, _ := strconv.Atoi(r.PostFormValue("device"))
	dev_id := r.PostFormValue("dev_id")
	user := &auth.User{
		Id:        id,
		Email:     email,
		Device:    device,
		DevId:     dev_id,
		FirstName: firstname,
		LastName:  lastname,
	}

	at := auth.NewAuthToken(*user, time.Now(), secret)
	//TODO store to redis uset and at

	fmt.Println("Saving to Redis")

	storage.SaveUser(*user)
	storage.SaveAuthToken(*at)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	rec := RegisterStatus{
		Token: at.HMAC,
	}
	if json.NewEncoder(w).Encode(rec) != nil {
		w.WriteHeader(500)
	}
}

func testNotificationSending(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, error := authenticate(r.Header.Get("auth_token"))
	if error != nil {
		w.WriteHeader(400)
		return
	}
	notificationSender.Send([]auth.User{*user}, storage.Vote{Name: "Hello world"})

	w.WriteHeader(200)
}

func emailVote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	token := r.FormValue("token")
	_, error := authenticate(token)
	if error != nil {
		w.WriteHeader(400)
		return
	}
	id := r.PostFormValue("vote")
	vote := storage.Vote{
		Id:   id,
		Name: "debug",
	}
	value, _ := strconv.Atoi(r.PostFormValue("value"))
	res := DoVoteStatus{
		Vote: DoVote{
			Name:  vote.Name,
			Value: value,
		},
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if json.NewEncoder(w).Encode(res) != nil {
		w.WriteHeader(500)
	}
}
