package notification

import (
	"bytes"
	"fmt"
	"time"

	"github.com/alexjlockwood/gcm"
	"github.com/mostafah/mandrill"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
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
	GoogleIds []string
	AppleIds  []string
	Emails    []string
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

func (this *Sender) Send(users []auth.User, message string) error {

	var buffer bytes.Buffer
	var devices Devices

	for i := range users {
		switch users[i].Device {
		case DEVICE_IOS:
			devices.AppleIds = append(devices.AppleIds, users[i].DevId)
		case DEVICE_ANDROID:
			devices.GoogleIds = append(devices.GoogleIds, users[i].DevId)
		default:
			devices.Emails = append(devices.Emails, users[i].Email)
		}
	}

	if len(devices.GoogleIds) > 0 {
		data := map[string]interface{}{"message": message} //TODO
		msg := gcm.NewMessage(data, devices.GoogleIds...)
		sender := &gcm.Sender{ApiKey: this.GoogleApiKey}
		_, err := sender.Send(msg, 2)
		if err != nil {
			buffer.WriteString(fmt.Sprintf("Sending notification to Google device failed: %s\n", err.Error()))
		}
	}

	if len(devices.AppleIds) > 0 {
		apn, err := apns.New(this.AppleCertFilename, this.AppleKeyFilename, this.AppleServer, 1*time.Second)
		if err != nil {
			buffer.WriteString(fmt.Sprintf("Sending notification to Apple device failed: %s\n", err.Error()))
		} else {
			for i := range devices.AppleIds {
				payload := apns.Payload{}
				payload.Aps.Alert.Body = message //TODO
				notification := apns.Notification{}
				notification.DeviceToken = devices.AppleIds[i]
				notification.Identifier = 0
				notification.Payload = &payload
				err = apn.Send(&notification)
				if err != nil {
					buffer.WriteString(fmt.Sprintf("Sending notification to Apple device failed: %s\n", err.Error()))
				}
			}
			apn.Close()
		}
	}

	if len(devices.Emails) > 0 {
		mandrill.Key = this.MandrillKey
		err := mandrill.Ping()
		if err != nil {
			buffer.WriteString(fmt.Sprintf("Sending notification to Email failed: %s\n", err.Error()))
		} else {
			data := make(map[string]string)
			data["QUESTION"] = "Test question"
			data["VOTE"] = "Test vote"
			data["TOKEN"] = "Test token"
			for i := range devices.Emails {
				msg := mandrill.NewMessageTo(devices.Emails[i], "")
				msg.Subject = this.MandrillSubject
				msg.FromEmail = this.MandrillFromEmail
				msg.FromName = this.MandrillFromName
				_, err := msg.SendTemplate(this.MandrillTemplate, data, false)
				if err != nil {
					buffer.WriteString(fmt.Sprintf("Sending notification to Email failed: %s\n", err.Error()))
				}
			}

		}
	}
	if len(buffer.String()) > 0 {
		return fmt.Errorf(buffer.String())
	} else {
		return nil
	}

}
