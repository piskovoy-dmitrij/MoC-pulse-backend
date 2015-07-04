package notification

import (
	"fmt"
	"time"

	"github.com/alexjlockwood/gcm"
	"github.com/mostafah/mandrill"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/storage"
	"github.com/virushuo/Go-Apns"
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
		apn, err := apns.New(this.AppleCertFilename, this.AppleKeyFilename, this.AppleServer, 1*time.Second)
		if err != nil {
			fmt.Printf("Notification sender ERROR: Sending notification to Apple device failed: %s\n", err.Error())
		} else {
			for i := range devices.AppleIds {
				fmt.Printf("Notification sender debug: Trying to send Apple device notifications to %v", devices.AppleIds[i])
				payload := apns.Payload{}
				payload.Aps.Alert.Body = vote.Name //TODO
				notification := apns.Notification{}
				notification.DeviceToken = devices.AppleIds[i]
				notification.Identifier = 0
				notification.Payload = &payload
				err = apn.Send(&notification)
				if err != nil {
					fmt.Printf("Notification sender ERROR: Sending notification to Apple device failed: %s\n", err.Error())
				}
			}
			apn.Close()
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
