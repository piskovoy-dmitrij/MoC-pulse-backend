package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/piskovoy-dmitrij/MoC-pulse-backend/Godeps/_workspace/src/github.com/FogCreek/mini"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/Godeps/_workspace/src/github.com/julienschmidt/httprouter"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/notification"
	"github.com/piskovoy-dmitrij/MoC-pulse-backend/tcpsocket"
)

var notificationSender *notification.Sender

var cfg *mini.Config

const cfgPath string = ".pulseconfigrc"

var Error *log.Logger
var Warning *log.Logger
var Info *log.Logger
var Debug *log.Logger

func main() {

	var err error
	var logLevel int64
	var logPath string

	cfg, err = mini.LoadConfiguration(cfgPath)
	if err != nil {
		log.Fatalf("Failed to open configuration file %s!\n", cfgPath)
	}

	Error = log.New(nil, "ERROR: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	Warning = log.New(nil, "WARNING: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	Info = log.New(nil, "INFO: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	Debug = log.New(nil, "DEBUG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	logLevel = cfg.Integer("LogLevel", 0)
	if logLevel >= 1 {
		logPath = cfg.String("LogPath", "MoC-pulse-backend.log")
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file %s: %v\n", logPath, err)
		}
		defer file.Close()
		Error = log.New(file, "ERROR: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
		if logLevel >= 2 {
			Warning = log.New(file, "WARNING: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
			if logLevel >= 3 {
				Info = log.New(file, "INFO: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
				if logLevel >= 4 {
					Debug = log.New(file, "DEBUG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
				}
			}
		}
	}

	defer Info.Printf("MoC-pulse-backend server app stopped.\n")
	Info.Printf("MoC-pulse-backend server app starting...\n")
	Debug.Printf("MoC-pulse-backend server app initialization: start\n")

	Debug.Printf("Initializing notification sender...\n")
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

	Debug.Printf("Initializing httprouter...\n")
	router := httprouter.New()
	router.GET("/votes", getVotes)
	router.POST("/votes", createVote)
	router.GET("/votes/:id", getVote)
	router.PUT("/votes/:id", doVote)
	router.GET("/vote", emailVote)
	router.POST("/user", registerUser)
	router.POST("/test_ios_notification", testIOSNotificationSending)
	router.POST("/test_android_notification", testAndroidNotificationSending)

	Debug.Printf("MoC-pulse-backend server app initialization: end\n")

	routerPort := cfg.Integer("HttpRouterPort", 3001)
	Debug.Printf("Starting httprouter on port %d as goroutine...\n", routerPort)
	go http.ListenAndServe(":"+strconv.FormatInt(routerPort, 10), router)

	tcpsocketPort := cfg.Integer("TcpSocketPort", 4242)
	Debug.Printf("Starting tcpsocket server on port %d...\n", tcpsocketPort)
	tcpsocket.ListenAndServer(":" + strconv.FormatInt(tcpsocketPort, 10))

	Info.Printf("MoC-pulse-backend server app started.\n")
}
