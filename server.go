package main

import (
	"flag"
	"net/http"

	logger "github.com/Sirupsen/logrus"
	"golang.org/x/net/websocket"

	"github.com/alittlebrighter/switchboard/appContext"
	"github.com/alittlebrighter/switchboard/persistence"
)

func main() {
	host := flag.String("host", "localhost:12345", "The relay host to connect to.")
	debug := flag.Bool("debug", false, "Sets the logging level to DEBUG.")
	flag.Parse()

	log := logger.WithFields(logger.Fields{
		"func": "main",
	})

	logger.SetLevel(logger.WarnLevel)
	if *debug {
		logger.SetLevel(logger.DebugLevel)
		log.Debug("Logging level set to DebugLevel.")
	}

	serverCtx := appContext.NewServerContext(persistence.MapBackend)

	http.Handle("/socket", websocket.Handler(serverCtx.WebsocketConn))
	http.HandleFunc("/messages", serverCtx.HTTPConn)

	log.WithFields(logger.Fields{
		"backend": "map",
		"host":    *host,
	}).Debugln("Starting server.")
	err := http.ListenAndServe(*host, nil)
	if err != nil {
		log.WithError(err).Fatalln("Server encountered an error.")
	}
}
