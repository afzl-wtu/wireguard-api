package main

import (
	"embed"
	"io/fs"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/afzl-wtu/wireguard-api/api"
	"github.com/afzl-wtu/wireguard-api/interfaces"
	"github.com/afzl-wtu/wireguard-api/store"
	"github.com/afzl-wtu/wireguard-api/utils"
)

//go:embed assets/*
var embeddedAssets embed.FS

func main() {
	store, err := store.New("./db")
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Init(); err != nil {
		log.Fatal(err)
	}
	assets, _ := fs.Sub(fs.FS(embeddedAssets), "templates")
	initServerConfig(store, assets)
	apiServer := api.NewApiServer()
	// assetsDir, _ := fs.Sub(fs.FS(embeddedAssets), "assets")
	log.Info("Starting server on", apiServer.Addr)
	if err := apiServer.Start(store); err != nil {
		log.Fatal(err)
	}
}

func initServerConfig(db interfaces.IStore, assetsDir fs.FS) {
	settings, err := db.GetGlobalSettings()
	if err != nil {
		log.Fatalf("Cannot get global settings: %v", err)
	}

	if _, err := os.Stat(settings.ConfigFilePath); err == nil {
		// file exists, don't overwrite it implicitly
		return
	}

	server, err := db.GetServer()
	if err != nil {
		log.Fatalf("Cannot get server config: %v", err)
	}

	clients, err := db.GetClients(false)
	if err != nil {
		log.Fatalf("Cannot get client config: %v", err)
	}

	// write config file
	err = utils.WriteWireGuardServerConfig(assetsDir, server, clients, settings)
	if err != nil {
		log.Fatalf("Cannot create server config: %v", err)
	}
}
