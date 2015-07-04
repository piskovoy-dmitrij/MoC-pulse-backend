package main

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"net/http"
	"strconv"
	"time"
)

var secret string = "shjgfshfkjgskdfjgksfghks"

type RegisterStatus struct {
	Token string `json: token`
}

func authenticate(r *http.Request) auth.User, error{
	
}

func createVote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, error := authenticate(r)
}

func getVote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, error := authenticate(r)
}

func getVotes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, error := authenticate(r)
}

func doVote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, error := authenticate(r)
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
