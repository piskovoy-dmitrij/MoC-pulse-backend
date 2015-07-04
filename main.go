package main

import (
	"fmt"
	"github.com/FogCreek/mini"
	"github.com/julienschmidt/httprouter"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/notification"
	"net/http"
)

var notificationSender *notification.Sender

func fatal(v interface{}) {
	fmt.Println(v)
}

func chk(err error) {
	if err != nil {
		fatal(err)
	}
}

func params() string {
	cfg, err := mini.LoadConfiguration(".pulseconfigrc")

	chk(err)

	info := fmt.Sprintf("db=%s",
		cfg.String("db", "127.0.0.1"),
	)
	return info
}

func main() {
	notificationSender = notification.NewSender("", "", "", "", "", "", "", "", "")
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
