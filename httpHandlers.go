package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/events"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/log"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/storage"
)

type RegisterStatus struct {
	Token string `json: token`
}

type ParamSt struct {
	Name string `json: name`
	Type string `json: type`
}

type DoVotePrm struct {
	Value int `json: value`
}

func createVote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	funcPrefix := "New vote creation"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	storageConnection := storage.NewStorageConnection(dbConnectionAddress)
	defer storageConnection.CloseStorageConnection()

	log.Debug.Printf("%s: authenticating user...\n", funcPrefix)
	user, error := storageConnection.Authenticate(r.Header.Get("auth_token"))
	if error != nil {
		log.Error.Printf("%s: user authentication failed\n", funcPrefix)
		w.WriteHeader(400)
		return
	}

	var params ParamSt
	log.Debug.Printf("%s: decoding params...\n", funcPrefix)
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		log.Error.Printf("%s: decoding params failed: %s\n", funcPrefix, err.Error())
		w.WriteHeader(400)
		return
	}

	log.Debug.Printf("%s: adding new vote to storage...\n", funcPrefix)
	vote, err := storageConnection.NewVote(params.Name, user.Id)
	if err != nil {
		log.Error.Printf("%s: adding vote '%s' to storage failed: %s\n", funcPrefix, params.Name, err.Error())
		return
	}

	log.Debug.Printf("%s: getting vote result status...\n", funcPrefix)
	res, err := storageConnection.GetVoteResultStatus(*vote, *user, storageConnection.UsersCount())
	if err != nil {
		log.Error.Printf("%s: getting vote result status failed: %s\n", funcPrefix, err.Error())
		w.WriteHeader(400)
		return
	}

	log.Debug.Printf("%s: getting users from storage...\n", funcPrefix)
	users, _ := storageConnection.GetUsers()
	log.Debug.Printf("%s: removing vote creator from notification list...\n", funcPrefix)
	for p, v := range users {
		if user.Id == v.Id {
			users = append(users[:p], users[p+1:]...)
			log.Debug.Printf("%s: vote creator has been found and succesfully removed from the list\n", funcPrefix)
			break
		}
	}
	log.Debug.Printf("%s: sending notifications to users...\n", funcPrefix)
	notificationSender.Send(users, res, dbConnectionAddress)

	log.Debug.Printf("%s: sending new vote event...\n", funcPrefix)
	*events.GetNewVoteChan() <- events.NewVoteEvent{Vote: res}

	log.Info.Printf("%s: new vote '%s' has been succesfully created!\n", funcPrefix, params.Name)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if json.NewEncoder(w).Encode(res) != nil {
		log.Error.Printf("%s: encoding response failed\n", funcPrefix)
		w.WriteHeader(500)
	}
}

func getVote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	funcPrefix := "Getting vote results"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	storageConnection := storage.NewStorageConnection(dbConnectionAddress)
	defer storageConnection.CloseStorageConnection()

	log.Debug.Printf("%s: authenticating user...\n", funcPrefix)
	user, error := storageConnection.Authenticate(r.Header.Get("auth_token"))
	if error != nil {
		log.Error.Printf("%s: user authentication failed\n", funcPrefix)
		w.WriteHeader(400)
		return
	}
	id := ps.ByName("id")

	log.Debug.Printf("%s: getting vote with id '%s' from storage...\n", funcPrefix, id)
	vote, err := storageConnection.GetVote(id)
	if err != nil {
		log.Error.Printf("%s: getting vote with id '%s' from storage failed: %s\n", funcPrefix, id, err.Error())
		w.WriteHeader(400)
		return
	}

	log.Info.Printf("%s: vote was successfully found: [%+v]\n", funcPrefix, vote)

	log.Debug.Printf("%s: getting vote result status...\n", funcPrefix)
	res, err := storageConnection.GetVoteResultStatus(*vote, *user, storageConnection.UsersCount())
	if err != nil {
		log.Error.Printf("%s: getting vote result status failed: %s\n", funcPrefix, err.Error())
		w.WriteHeader(400)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if json.NewEncoder(w).Encode(res) != nil {
		log.Error.Printf("%s: encoding response failed\n", funcPrefix)
		w.WriteHeader(500)
	}
}

func getVotes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	funcPrefix := "Getting all votes with results"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	storageConnection := storage.NewStorageConnection(dbConnectionAddress)
	defer storageConnection.CloseStorageConnection()

	log.Debug.Printf("%s: authenticating user...\n", funcPrefix)
	user, error := storageConnection.Authenticate(r.Header.Get("auth_token"))
	if error != nil {
		log.Error.Printf("%s: user authentication failed\n", funcPrefix)
		w.WriteHeader(400)
		return
	}
	log.Debug.Printf("%s: getting all votes with results from storage...\n", funcPrefix)
	votes, err := storageConnection.GetAllVotesWithResult(*user)
	if err != nil {
		log.Error.Printf("%s: getting all votes with results from storage failed: %s\n", funcPrefix, err.Error())
		w.WriteHeader(400)
		return
	}
	res := storage.VotesStatus{
		Votes: votes,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if json.NewEncoder(w).Encode(res) != nil {
		log.Error.Printf("%s: encoding response failed\n", funcPrefix)
		w.WriteHeader(500)
	}
}

func doVote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	funcPrefix := "Processing voting"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	storageConnection := storage.NewStorageConnection(dbConnectionAddress)
	defer storageConnection.CloseStorageConnection()

	log.Debug.Printf("%s: authenticating user...\n", funcPrefix)
	user, error := storageConnection.Authenticate(r.Header.Get("auth_token"))
	if error != nil {
		log.Error.Printf("%s: user authentication failed\n", funcPrefix)
		w.WriteHeader(400)
		return
	}
	id := ps.ByName("id")

	log.Debug.Printf("%s: getting vote with id '%s' from storage...\n", funcPrefix, id)
	vote, err := storageConnection.GetVote(id)
	if err != nil {
		log.Error.Printf("%s: getting vote with id '%s' from storage failed: %s\n", funcPrefix, id, err.Error())
		w.WriteHeader(400)
		return
	}

	if storageConnection.IsVotedByUser(*vote, *user) {
		log.Warning.Printf("%s: user has already voted!\n", funcPrefix)
		w.WriteHeader(200)
		return
	}

	var params DoVotePrm
	log.Debug.Printf("%s: decoding params...\n", funcPrefix)
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		log.Error.Printf("%s: decoding params failed: %s\n", funcPrefix, err.Error())
		w.WriteHeader(400)
		return
	}

	log.Debug.Printf("%s: modifying vote...\n", funcPrefix)
	storageConnection.VoteProcessing(*vote, *user, params.Value)

	log.Debug.Printf("%s: getting vote result status...\n", funcPrefix)
	res, err := storageConnection.GetVoteResultStatus(*vote, *user, storageConnection.UsersCount())
	if err != nil {
		log.Error.Printf("%s: getting vote result status failed: %s\n", funcPrefix, err.Error())
		w.WriteHeader(400)
		return
	}

	log.Debug.Printf("%s: sending vote update event...\n", funcPrefix)
	*events.GetVoteUpdateChan() <- events.VoteUpdateEvent{Vote: res}

	log.Info.Printf("%s: vote '%s' has been succesfully updated!\n", funcPrefix, vote.Name)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if json.NewEncoder(w).Encode(res) != nil {
		log.Error.Printf("%s: encoding response failed\n", funcPrefix)
		w.WriteHeader(500)
	}
}

func registerUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	funcPrefix := "Registering new user"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	storageConnection := storage.NewStorageConnection(dbConnectionAddress)
	defer storageConnection.CloseStorageConnection()

	log.Debug.Printf("%s: getting params from request...\n", funcPrefix)
	id := r.PostFormValue("id")
	email := r.PostFormValue("email")
	firstname := r.PostFormValue("first_name")
	lastname := r.PostFormValue("last_name")
	device, _ := strconv.Atoi(r.PostFormValue("device"))
	dev_id := r.PostFormValue("dev_id")
	token := r.PostFormValue("token")

	if token != "BE7C411D475AEA4CF1D7B472D5BD1" {
		log.Warning.Printf("%s: token is not correct!\n", funcPrefix)
		w.WriteHeader(403)
		return
	}

	log.Debug.Printf("%s: checking existence of user with id '%s'...\n", funcPrefix, id)
	user, err := storageConnection.LoadUser("user:" + id)
	if err == nil {
		log.Debug.Printf("%s: user with id '%s' already exists; modifying his params...\n", funcPrefix, id)
		user.Device = device
		user.DevId = dev_id
		user.Email = email
		user.FirstName = firstname
		user.LastName = lastname
	} else {
		log.Debug.Printf("%s: user with id '%s' was not found; creating new user...\n", funcPrefix, id)
		user = &auth.User{
			Id:        id,
			Email:     email,
			Device:    device,
			DevId:     dev_id,
			FirstName: firstname,
			LastName:  lastname,
		}
	}
	log.Debug.Printf("%s: composing new auth token for user with id '%s'...\n", funcPrefix, id)
	at := auth.NewAuthToken(*user, time.Now())

	log.Debug.Printf("%s: saving user with id '%s' to storage...\n", funcPrefix, id)
	storageConnection.SaveUser(*user)

	storageConnection.SaveAuthToken(*at)

	log.Info.Printf("%s: user has been succesfully registered: [%+v]\n", funcPrefix, user)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	rec := RegisterStatus{
		Token: at.HMAC,
	}
	if json.NewEncoder(w).Encode(rec) != nil {
		log.Error.Printf("%s: encoding response failed\n", funcPrefix)
		w.WriteHeader(500)
	}
}

func testIOSNotificationSending(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	funcPrefix := "iOS notification sending test"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	log.Debug.Printf("%s: getting dev_id from request...\n", funcPrefix)
	dev_id := r.PostFormValue("dev_id")
	if dev_id != "" {
		log.Info.Printf("%s: trying to send notification for dev_id '%s'...\n", funcPrefix, dev_id)
		user := &auth.User{Id: "100", FirstName: "John", LastName: "Doe", Device: 0, DevId: dev_id}
		result := &storage.Result{Yellow: 2, Green: 1, Red: 3, AllUsers: 6, VoteUsers: 6}
		voteWithResult := &storage.VoteWithResult{Id: "5", Name: "HelloWorld", Date: 1436966974, Voted: true, Owner: *user, Result: *result}
		voteResultStatus := &storage.VoteResultStatus{Vote: *voteWithResult}
		notificationSender.Send([]auth.User{*user}, voteResultStatus, dbConnectionAddress)
		w.WriteHeader(200)
	} else {
		log.Warning.Printf("%s: there is no dev_id in request!\n", funcPrefix)
		w.WriteHeader(400)
	}
}

func testAndroidNotificationSending(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	funcPrefix := "Android notification sending test"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	log.Debug.Printf("%s: getting dev_id from request...\n", funcPrefix)
	dev_id := r.PostFormValue("dev_id")
	if dev_id != "" {
		log.Info.Printf("%s: trying to send notification for dev_id '%s'...\n", funcPrefix, dev_id)
		user := &auth.User{Id: "100", FirstName: "John", LastName: "Doe", Device: 1, DevId: dev_id}
		result := &storage.Result{Yellow: 2, Green: 1, Red: 3, AllUsers: 6, VoteUsers: 6}
		voteWithResult := &storage.VoteWithResult{Id: "5", Name: "HelloWorld", Date: 1436966974, Voted: true, Owner: *user, Result: *result}
		voteResultStatus := &storage.VoteResultStatus{Vote: *voteWithResult}
		notificationSender.Send([]auth.User{*user}, voteResultStatus, dbConnectionAddress)
		w.WriteHeader(200)
	} else {
		log.Warning.Printf("%s: there is no dev_id in request!\n", funcPrefix)
		w.WriteHeader(400)
	}
}

func emailVote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	funcPrefix := "Processing email voting"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	storageConnection := storage.NewStorageConnection(dbConnectionAddress)
	defer storageConnection.CloseStorageConnection()

	token := r.FormValue("token")
	log.Debug.Printf("%s: authenticating by token '%s' from request...\n", funcPrefix, token)
	_, error := storageConnection.Authenticate(token)
	if error != nil {
		log.Error.Printf("%s: authentication failed\n", funcPrefix)
		w.WriteHeader(400)
		return
	}
	id := r.PostFormValue("vote")
	log.Debug.Printf("%s: getting vote with id '%s' from storage...\n", funcPrefix, id)
	vote, _ := storageConnection.GetVote(id)
	value, _ := strconv.Atoi(r.PostFormValue("value"))
	log.Debug.Printf("%s: modifying vote...\n", funcPrefix)
	res := storage.DoVoteStatus{
		Vote: storage.DoVote{
			Name:  vote.Name,
			Value: value,
		},
	}

	log.Info.Printf("%s: vote '%s' has been succesfully updated!\n", funcPrefix, vote.Name)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if json.NewEncoder(w).Encode(res) != nil {
		log.Error.Printf("%s: encoding response failed\n", funcPrefix)
		w.WriteHeader(500)
	}
}
