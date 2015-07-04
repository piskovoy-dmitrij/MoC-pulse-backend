package main

import (
	"encoding/json"
	//	"errors"
	"github.com/julienschmidt/httprouter"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/redis"
	"net/http"
	"strconv"
	"time"
)

var secret string = "shjgfshfkjgskdfjgksfghks"

type RegisterStatus struct {
	Token string `json:"token"`
}

type VoteStatus struct {
	Vote redis.Vote `json:"vote"`
}

func authenticate(r *http.Request) (*auth.User, error) {
	token := r.Header.Get("token")
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
	user, error := authenticate(r)
	if error != nil {
		w.WriteHeader(400)
		return
	}
	name := r.PostFormValue("name")
	vote := redis.NewVote(name, user.Id)
	res := VoteStatus{
		Vote: *vote,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if json.NewEncoder(w).Encode(res) != nil {
		w.WriteHeader(500)
	}
}

func getVote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, error := authenticate(r)
	if error != nil {
		w.WriteHeader(400)
		return
	}
	if json.NewEncoder(w).Encode(user) != nil {
		w.WriteHeader(500)
	}
}

func getVotes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, error := authenticate(r)
	if error != nil {
		w.WriteHeader(400)
		return
	}
	if json.NewEncoder(w).Encode(user) != nil {
		w.WriteHeader(500)
	}
}

func doVote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, error := authenticate(r)
	if error != nil {
		w.WriteHeader(400)
		return
	}
	if json.NewEncoder(w).Encode(user) != nil {
		w.WriteHeader(500)
	}
}

func registerUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id := r.PostFormValue("id")
	email := r.PostFormValue("email")
	device, _ := strconv.Atoi(r.PostFormValue("device"))
	dev_id := r.PostFormValue("dev_id")
	user := &auth.User{
		Id:     id,
		Email:  email,
		Device: device,
		DevId:  dev_id,
	}

	at := auth.NewAuthToken(*user, time.Now(), secret)
	//TODO store to redis uset and at
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	rec := RegisterStatus{
		Token: at.HMAC,
	}
	if json.NewEncoder(w).Encode(rec) != nil {
		w.WriteHeader(500)
	}
}
