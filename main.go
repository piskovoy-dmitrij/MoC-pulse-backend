package main

import (
	"fmt"

	"github.com/FogCreek/mini"
	"github.com/julienschmidt/httprouter"

	"net"
	"net/http"
	"os"

	"github.com/walkline/MoC-pulse-backend/notification"
	"github.com/walkline/MoC-pulse-backend/tcpsocket"
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

	println("Starting http server...")

	// starting new goroutine
	go http.ListenAndServe(":8080", router)

	l, err := net.Listen("tcp", ":4242")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	println("Starting tcp server...")
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on :4242...")
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go tcpsocket.HandleNewConnection(conn)
	}
}
