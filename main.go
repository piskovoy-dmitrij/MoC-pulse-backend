package main

import (
	syslog "log"
	"net/http"
	"strconv"

	"github.com/FogCreek/mini"
	"github.com/julienschmidt/httprouter"
	log "github.com/piskovoy-dmitrij/MoC-pulse-backend/log"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/notification"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/tcpsocket"
)

var notificationSender *notification.Sender

var cfg *mini.Config

const cfgPath string = ".pulseconfigrc"

func main() {

	var err error

	cfg, err = mini.LoadConfiguration(cfgPath)
	if err != nil {
		syslog.Fatalf("Failed to open configuration file %s!\n", cfgPath)
	}

	log.NewLogger(
		cfg.Integer("LogLevel", 0),
		cfg.String("LogPath", "MoC-pulse-backend.log"))
	defer log.CloseLog()
	defer log.Info.Printf("MoC-pulse-backend server app stopped.\n")
	log.Info.Printf("MoC-pulse-backend server app starting...\n")
	log.Debug.Printf("MoC-pulse-backend server app initialization: start\n")

	log.Debug.Printf("Initializing notification sender...\n")
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

	log.Debug.Printf("Initializing httprouter...\n")
	router := httprouter.New()
	router.GET("/votes", getVotes)
	router.POST("/votes", createVote)
	router.GET("/votes/:id", getVote)
	router.PUT("/votes/:id", doVote)
	router.GET("/vote", emailVote)
	router.POST("/user", registerUser)
	router.POST("/test_ios_notification", testIOSNotificationSending)
	router.POST("/test_android_notification", testAndroidNotificationSending)

	log.Debug.Printf("MoC-pulse-backend server app initialization: end\n")

	routerPort := cfg.Integer("HttpRouterPort", 3001)
	log.Debug.Printf("Starting httprouter on port %d as goroutine...\n", routerPort)
	go http.ListenAndServe(":"+strconv.FormatInt(routerPort, 10), router)

	tcpsocketPort := cfg.Integer("TcpSocketPort", 4242)
	log.Debug.Printf("Starting tcpsocket server on port %d...\n", tcpsocketPort)
	log.Info.Printf("MoC-pulse-backend server app started.\n")
	tcpsocket.ListenAndServer(":"+strconv.FormatInt(tcpsocketPort, 10), notificationSender)
}
