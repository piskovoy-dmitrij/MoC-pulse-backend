package notification

import (
	"encoding/json"
	"fmt"
	"github.com/Mistobaan/go-apns"
	"github.com/alexjlockwood/gcm"
	"github.com/mostafah/mandrill"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/storage"
)

type Sender struct {
	GoogleApiKey      string
	AppleCertFilename string
	AppleKeyFilename  string
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
	Alert    string `json:"alert"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Id       string `json:"id"`
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
	AppleCertFilename string,
	AppleKeyFilename string,
	AppleServer string,
	MandrillKey string,
	MandrillTemplate string,
	MandrillFromEmail string,
	MandrillFromName string,
	MandrillSubject string) *Sender {

	return &Sender{
		GoogleApiKey,
		AppleCertFilename,
		AppleKeyFilename,
		AppleServer,
		MandrillKey,
		MandrillTemplate,
		MandrillFromEmail,
		MandrillFromName,
		MandrillSubject,
	}
}

func (this *Sender) Send(users []auth.User, vote storage.Vote) {
	go this.send(users, vote)
}

func (this *Sender) send(users []auth.User, vote storage.Vote) {

	var devices Devices

	for i := range users {
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
		fmt.Printf("Notification sender debug: Trying to send Google device notifications to %v", devices.GoogleIds)
		data := map[string]interface{}{"message": vote.Name} //TODO
		msg := gcm.NewMessage(data, devices.GoogleIds...)
		sender := &gcm.Sender{ApiKey: this.GoogleApiKey}
		_, err := sender.Send(msg, 2)
		if err != nil {
			fmt.Printf("Notification sender ERROR: Sending notification to Google device failed: %s\n", err.Error())
		}
	}

	if len(devices.AppleIds) > 0 {
		apn, err := apns.NewClient(this.AppleServer, this.AppleCertFilename, this.AppleKeyFilename)
		if err != nil {
			fmt.Printf("Notification sender ERROR: Sending notification to Apple device failed: %s\n", err.Error())
		} else {
			payload := &ApplePayload{}
			payload.Aps.Alert = vote.Name
			payload.Aps.Title = "MOC Pulse"
			payload.Aps.Category = "watchkit"
			payload.Aps.Id = vote.Id
			actions := &SimulatorAction{}
			actions.Title = "Vote"
			actions.Identifier = "voteButtonAction"
			payload.SimulatorActions = append(payload.SimulatorActions, *actions)
			bytes, _ := json.Marshal(payload)
			fmt.Printf("Notification sender debug: %v", string(bytes))
			for i := range devices.AppleIds {
				fmt.Printf("Notification sender debug: Trying to send Apple device notifications to %v", devices.AppleIds[i])
				err = apn.SendPayloadString(devices.AppleIds[i], bytes, 5)
				if err != nil {
					fmt.Printf("Notification sender ERROR: Sending notification to Apple device failed: %s\n", err.Error())
				}
			}
		}
	}

	if len(devices.OtherUsers) > 0 {
		mandrill.Key = this.MandrillKey
		err := mandrill.Ping()
		if err != nil {
			fmt.Printf("Notification sender ERROR: Sending notification to Email failed: %s\n", err.Error())
		} else {
			data := make(map[string]string)
			data["QUESTION"] = vote.Name
			data["VOTE"] = vote.Id
			for i := range devices.OtherUsers {
				data["TOKEN"] = devices.OtherUsers[i].Id
				fmt.Printf("Notification sender debug: Trying to send Email notifications to %v", devices.OtherUsers[i].Email)
				msg := mandrill.NewMessageTo(devices.OtherUsers[i].Email, devices.OtherUsers[i].FirstName+devices.OtherUsers[i].LastName)
				msg.Subject = this.MandrillSubject
				msg.FromEmail = this.MandrillFromEmail
				msg.FromName = this.MandrillFromName
				_, err := msg.SendTemplate(this.MandrillTemplate, data, false)
				if err != nil {
					fmt.Printf("Notification sender ERROR: Sending notification to Email failed: %s\n", err.Error())
				}
			}
		}
	}
}
