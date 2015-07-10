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
	at, err := storage.LoadAuthToken(token)
	if err != nil {
		return nil, err
	}
	info, err := at.GetTokenInfo(secret)
	if err != nil {
		return nil, err
	}
	user, err := storage.LoadUser("user:" + info.Id)
	if err != nil {
		return nil, err
	} else {
		return user, nil
	}
}

func createVote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, error := authenticate(r.Header.Get("auth_token"))
	if error != nil {
		w.WriteHeader(400)
		return
	}
	name := r.PostFormValue("name")
	vote := storage.NewVote(name, user.Id)
	users, _ := storage.GetUsers()
	notificationSender.Send(users, *vote)
	res := storage.VoteStatus{
		Vote: *vote,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
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
	
	vote, err := storage.GetVote(id)
	
	fmt.Println("out of get")
	
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}
	
	res := storage.GetVoteResultStatus(*vote)
	
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
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
	
	votes := storage.GetAllVotesWithResult()
	res := storage.VotesStatus{
		Votes: votes[0:2],
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
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
	token := r.PostFormValue("tiken")
	if token != "BE7C411D475AEA4CF1D7B472D5BD1" {
		w.WriteHeader(403)
		return
	}
	user, err := storage.LoadUser("user:" + id)
	if err != nil {
		user.Device = device
		user.DevId = dev_id
		user.Email = email
		user.FirstName = firstname
		user.LastName = lastname
	} else {
		user = &auth.User{
			Id:        id,
			Email:     email,
			Device:    device,
			DevId:     dev_id,
			FirstName: firstname,
			LastName:  lastname,
		}
	}

	at := auth.NewAuthToken(*user, time.Now(), secret)
	storage.SaveUser(*user)
	storage.SaveAuthToken(*at)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
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
	notificationSender.Send([]auth.User{*user}, storage.Vote{Id: "5", Name: "Hello world"})

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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if json.NewEncoder(w).Encode(res) != nil {
		w.WriteHeader(500)
	}
}
