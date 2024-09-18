package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/afzl-wtu/wireguard-api/api"
)

func main() {
	apiServer := api.NewApiServer()
	log.Info("Starting server on", apiServer.Addr)
	if err := apiServer.Start(); err != nil {
		log.Fatal(err)
	}
}
