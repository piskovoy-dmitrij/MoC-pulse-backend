package notification

import (
	"bytes"
	"fmt"
	"net/smtp"
	"time"

	"github.com/alexjlockwood/gcm"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/auth"
	"github.com/virushuo/Go-Apns"
)

type Sender struct {
	GoogleApiKey      string
	AppleCertFilename string
	AppleKeyFilename  string
	AppleServer       string
	SmtpAuthUsername  string
	SmtpAuthPassword  string
	SmtpAuthHost      string
	SmtpServer        string
	SmtpSender        string
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
	SmtpAuthUsername string,
	SmtpAuthPassword string,
	SmtpAuthHost string,
	SmtpServer string,
	SmtpSender string) *Sender {

	return &Sender{
		GoogleApiKey,
		AppleCertFilename,
		AppleKeyFilename,
		AppleServer,
		SmtpAuthUsername,
		SmtpAuthPassword,
		SmtpAuthHost,
		SmtpServer,
		SmtpSender}
}

func (this *Sender) Send(users []auth.User, message string) error {

	var str string
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
			buffer.WriteString(fmt.Sprintf("Sending notification to Google device failed: %s\n", err))
		}
	}

	if len(devices.AppleIds) > 0 {
		apn, err := apns.New(this.AppleCertFilename, this.AppleKeyFilename, this.AppleServer, 1*time.Second)
		if err != nil {
			str = fmt.Sprintf("Sending notification to Apple device failed: %s\n", err)
			buffer.WriteString(str)
		} else {
			for i := range devices.AppleIds {
				payload := apns.Payload{}
				payload.Aps.Alert.Body = message
				notification := apns.Notification{}
				notification.DeviceToken = devices.AppleIds[i]
				notification.Identifier = 0
				notification.Payload = &payload
				err = apn.Send(&notification)
				if err != nil {
					buffer.WriteString(fmt.Sprintf("Sending notification to Apple device failed: %s\n", err))
				}
			}
			apn.Close()
		}
	}

	if len(devices.Emails) > 0 {
		auth := smtp.PlainAuth(
			"",
			this.SmtpAuthUsername,
			this.SmtpAuthPassword,
			this.SmtpAuthHost,
		)
		err := smtp.SendMail(
			this.SmtpServer,
			auth,
			this.SmtpSender,
			devices.Emails,
			[]byte(message),
		)
		if err != nil {
			buffer.WriteString(fmt.Sprintf("Sending notification to Email failed: %s\n", err))
		}
	}

	return fmt.Errorf(buffer.String())

}
