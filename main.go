package main

import (
	"fmt"

	"github.com/FogCreek/mini"
	"github.com/julienschmidt/httprouter"

	"net/http"

	"github.com/piskovoy-dmitrij/MoC-pulse-backend/notification"
)

var notificationSender *notification.Sender

var cfg *mini.Config

func fatal(v interface{}) {
	fmt.Println(v)
}

func chk(err error) {
	if err != nil {
		fatal(err)
	}
}

func main() {

	var err error

	cfg, err = mini.LoadConfiguration(".pulseconfigrc")
	chk(err)

	notificationSender = notification.NewSender(
		cfg.String("GoogleApiKey", ""),
		cfg.String("AppleCertPath", "pushcert.pem"),
		cfg.String("AppleKeyPath", "pushkey.pem"),
		cfg.String("AppleServer", "gateway.push.apple.com:2195"),
		cfg.String("MandrillKey", ""),
		cfg.String("MandrillTemplate", "vote"),
		cfg.String("MandrillFromEmail", "pulse@masterofcode.com"),
		cfg.String("MandrillFromName", "MoC Pulse"),
		cfg.String("MandrillSubject", "New Voting"))

	router := httprouter.New()
	router.GET("/votes", getVotes)
	router.POST("/votes", createVote)
	router.GET("/votes/:id", getVote)
	router.PUT("/votes/:id", doVote)
	router.GET("/vote", emailVote)
	router.POST("/user", registerUser)
	router.POST("/test_notification_sending", testNotificationSending)

	http.ListenAndServe(":8080", router)

}
