package notification

import (
	"encoding/json"

	"github.com/Mistobaan/go-apns"
	"github.com/alexjlockwood/gcm"
	"github.com/mostafah/mandrill"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/log"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/storage"
)

type Sender struct {
	GoogleApiKey      string
	AppleCertPath     string
	AppleKeyPath      string
	AppleServer       string
	MandrillKey       string
	MandrillTemplate  string
	MandrillFromEmail string
	MandrillFromName  string
	MandrillSubject   string
}

type Devices struct {
	GoogleIds  []string
	AppleIds   []string
	OtherUsers []auth.User /* for sending email */
}

type Aps struct {
	Alert    string                 `json:"alert"`
	Title    string                 `json:"title"`
	Category string                 `json:"category"`
	Vote     storage.VoteWithResult `json:"vote"`
}

type SimulatorAction struct {
	Title      string `json:"title"`
	Identifier string `json:"identifier"`
}

type ApplePayload struct {
	Aps              Aps               `json:"aps"`
	SimulatorActions []SimulatorAction `json:"WatchKit Simulator Actions"`
}

const (
	DEVICE_IOS = iota
	DEVICE_ANDROID
	DEVICE_OTHER /* for sending email */
)

func NewSender(GoogleApiKey string,
	AppleCertPath string,
	AppleKeyPath string,
	AppleServer string,
	MandrillKey string,
	MandrillTemplate string,
	MandrillFromEmail string,
	MandrillFromName string,
	MandrillSubject string) *Sender {

	return &Sender{
		GoogleApiKey,
		AppleCertPath,
		AppleKeyPath,
		AppleServer,
		MandrillKey,
		MandrillTemplate,
		MandrillFromEmail,
		MandrillFromName,
		MandrillSubject,
	}
}

func (this *Sender) Send(users []auth.User, vote storage.Vote, dbConnectionAddress string) {
	go this.send(users, vote, dbConnectionAddress)
}

func (this *Sender) send(users []auth.User, vote storage.Vote, dbConnectionAddress string) {
	funcPrefix := "Sending notifications"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	storageConnection := storage.NewStorageConnection(dbConnectionAddress)
	defer storageConnection.CloseStorageConnection()

	ownerUser, err := storageConnection.LoadUser("user:" + vote.Owner)
	if err != nil {
		log.Warning.Printf("%s: loading owner user from vote failed: %s\n", funcPrefix, err.Error())
		ownerUser = &auth.User{
			Id:        "",
			Email:     "",
			Device:    2,
			DevId:     "",
			FirstName: "",
			LastName:  "",
		}
	}

	// used for getting VoteWithResult structure from Vote
	res, error := storageConnection.GetVoteResultStatus(vote, *ownerUser)
	if error != nil {
		log.Error.Printf("%s: getting vote result status failed: %s\n", funcPrefix, err.Error())
	}
	voteWithResult := res.Vote

	var devices Devices

	log.Debug.Printf("%s: getting DevIds from users...\n", funcPrefix)
	for i := range users {
		log.Debug.Printf("%s: getting DevId from user %+v...\n", funcPrefix, users[i])
		switch users[i].Device {
		case DEVICE_IOS:
			devices.AppleIds = append(devices.AppleIds, users[i].DevId)
		case DEVICE_ANDROID:
			devices.GoogleIds = append(devices.GoogleIds, users[i].DevId)
		default:
			devices.OtherUsers = append(devices.OtherUsers, users[i])
		}
	}

	if len(devices.GoogleIds) > 0 {
		log.Debug.Printf("%s: trying to send Google device notifications to %v\n", funcPrefix, devices.GoogleIds)
		bytes, err := json.Marshal(voteWithResult)
		if err != nil {
			log.Error.Printf("%s: sending notification to Google device failed: %s\n", funcPrefix, err.Error())
		} else {
			log.Debug.Printf("%s: json for sending as notification to Google device: %v\n", funcPrefix, string(bytes))
			data := map[string]interface{}{"vote": string(bytes)}
			msg := gcm.NewMessage(data, devices.GoogleIds...)
			sender := &gcm.Sender{ApiKey: this.GoogleApiKey}
			_, err := sender.Send(msg, 2)
			if err != nil {
				log.Error.Printf("%s: sending notification to Google device failed: %s\n", funcPrefix, err.Error())
			}
		}
	}

	if len(devices.AppleIds) > 0 {
		log.Debug.Printf("%s: Apple notification Server: %s, Cert: %s, Key: %s\n", funcPrefix, this.AppleServer, this.AppleCertPath, this.AppleKeyPath)
		apn, err := apns.NewClient(this.AppleServer, this.AppleCertPath, this.AppleKeyPath)
		apn.MAX_PAYLOAD_SIZE = 2048
		if err != nil {
			log.Error.Printf("%s: sending notification to Apple device failed: %s\n", funcPrefix, err.Error())
		} else {
			payload := &ApplePayload{}
			payload.Aps.Alert = voteWithResult.Name
			payload.Aps.Title = "MOC Pulse"
			payload.Aps.Category = "newVote"
			payload.Aps.Vote = voteWithResult
			actions := &SimulatorAction{}
			actions.Title = "Vote"
			actions.Identifier = "voteButtonAction"
			payload.SimulatorActions = append(payload.SimulatorActions, *actions)
			bytes, _ := json.Marshal(payload)
			log.Debug.Printf("%s: Apple notification payload: %v\n", funcPrefix, string(bytes))
			for i := range devices.AppleIds {
				log.Debug.Printf("%s: trying to send Apple device notification to %v\n", funcPrefix, devices.AppleIds[i])
				err = apn.SendPayloadString(devices.AppleIds[i], bytes, 5)
				if err != nil {
					log.Error.Printf("%s: sending notification to Apple device failed: %s\n", funcPrefix, err.Error())
				}
			}
		}
	}

	if len(devices.OtherUsers) > 0 {
		mandrill.Key = this.MandrillKey
		err := mandrill.Ping()
		if err != nil {
			log.Error.Printf("%s: sending notification to Email failed: %s\n", funcPrefix, err.Error())
		} else {
			data := make(map[string]string)
			data["QUESTION"] = voteWithResult.Name
			data["VOTE"] = voteWithResult.Id
			for i := range devices.OtherUsers {
				data["TOKEN"] = devices.OtherUsers[i].Id
				log.Debug.Printf("%s: trying to send Email notification to %v\n", funcPrefix, devices.OtherUsers[i].Email)
				msg := mandrill.NewMessageTo(devices.OtherUsers[i].Email, devices.OtherUsers[i].FirstName+devices.OtherUsers[i].LastName)
				msg.Subject = this.MandrillSubject
				msg.FromEmail = this.MandrillFromEmail
				msg.FromName = this.MandrillFromName
				_, err := msg.SendTemplate(this.MandrillTemplate, data, false)
				if err != nil {
					log.Error.Printf("%s: sending notification to Email failed: %s\n", funcPrefix, err.Error())
				}
			}
		}
	}
}
