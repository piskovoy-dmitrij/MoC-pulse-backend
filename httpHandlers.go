package main

import (
	"encoding/json"
	"fmt"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/Godeps/_workspace/src/github.com/julienschmidt/httprouter"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/events"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/storage"
	"net/http"
	"strconv"
	"time"
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

	var params ParamSt
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		fmt.Println(err)
	}

	vote := storage.NewVote(params.Name, user.Id)

	// i think better use new goroutine
	go func() {
		users, _ := storage.GetUsers()
		notificationSender.Send(users, *vote)
	}()

	res := storage.GetVoteResultStatus(*vote, *user)

	*events.GetNewVoteChan() <- events.NewVoteEvent{res}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if json.NewEncoder(w).Encode(res) != nil {
		w.WriteHeader(500)
	}
}

func getVote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, error := authenticate(r.Header.Get("auth_token"))
	if error != nil {
		w.WriteHeader(400)
		return
	}
	id := ps.ByName("id")

	vote, err := storage.GetVote(id)

	if err != nil {
		w.WriteHeader(400)
		return
	}

	res := storage.GetVoteResultStatus(*vote, *user)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if json.NewEncoder(w).Encode(res) != nil {
		w.WriteHeader(500)
	}
}

func getVotes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, error := authenticate(r.Header.Get("auth_token"))
	if error != nil {
		w.WriteHeader(400)
		return
	}
	votes := storage.GetAllVotesWithResult(*user)
	res := storage.VotesStatus{
		Votes: votes,
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
	user, error := authenticate(r.Header.Get("auth_token"))
	if error != nil {
		w.WriteHeader(400)
		return
	}
	id := ps.ByName("id")
	vote, err := storage.GetVote(id)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}

	var params DoVotePrm
	errDec := json.NewDecoder(r.Body).Decode(&params)

	if errDec != nil {
		fmt.Println(errDec)
		w.WriteHeader(400)
		return
	}

	storage.VoteProcessing(*vote, *user, params.Value)

	res := storage.GetVoteResultStatus(*vote, *user)

	*events.GetVoteUpdateChan() <- events.VoteUpdateEvent{res}

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
	token := r.PostFormValue("token")
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
	/*user, error := authenticate(r.Header.Get("auth_token"))
	if error != nil {
		w.WriteHeader(400)
		return
	}*/
	//	notificationSender.Send([]auth.User{auth.User{Id: "100", FirstName: "John", LastName: "Doe", Device: 0, DevId: "ca4f2547a7fc19c4b92a27e940c373d3d3bded3102d5eddc4f63d74d615fab2c"}}, storage.Vote{Id: "5", Name: "Hello world"})
	notificationSender.Send([]auth.User{auth.User{Id: "100", FirstName: "John", LastName: "Doe", Device: 0, DevId: "14ad0aba799d831ad77239c7bb62b29e43bd7ebccb81716445b3124e42851ee8"}}, storage.Vote{Id: "5", Name: "Hello world"})

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
	vote, _ := storage.GetVote(id)
	value, _ := strconv.Atoi(r.PostFormValue("value"))
	res := storage.DoVoteStatus{
		Vote: storage.DoVote{
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
