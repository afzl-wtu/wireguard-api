package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"time"

	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/afzl-wtu/wireguard-api/api"
	"github.com/afzl-wtu/wireguard-api/interfaces"
	model "github.com/afzl-wtu/wireguard-api/models"
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
	assets, _ := fs.Sub(fs.FS(embeddedAssets), "assets")
	startServer := initServerConfig(store, assets)
	if !startServer {
		cmd := exec.Command("systemctl", "start", "wg-quick@wg0")

		// Run the command
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Info(string(output))
	}
	apiServer := api.NewApiServer()
	// assetsDir, _ := fs.Sub(fs.FS(embeddedAssets), "assets")
	log.Info("Starting server on", apiServer.Addr)
	if err := apiServer.Start(store); err != nil {
		log.Fatal(err)
	}
}

//	{
//	    "id": "crr9ub9mfo2s71cm18fg",
//	    "private_key": "eG6XZzelhlFCup7FgINMWb7glcZeXKFdjakFL2u1o1Y=",
//	    "public_key": "06lfJELdqtYi4nzWOTtxcExxtrfo0NXxeXO6uWLBIHM=",
//	    "preshared_key": "5nAHXBZd+WdcLBkclwBUTGITyPYV4FRvyvJF9Ee+Q3k=",
//	    "name": "delmee",
//	    "telegram_userid": "",
//	    "email": "",
//	    "allocated_ips": [
//	        "10.7.0.52/32"
//	    ],
//	    "allowed_ips": [
//	        "0.0.0.0/0"
//	    ],
//	    "extra_allowed_ips": [],
//	    "endpoint": "",
//	    "additional_notes": "",
//	    "use_server_dns": true,
//	    "enabled": true,
//	    "created_at": "2024-09-27T12:03:57.749946228Z",
//	    "updated_at": "2024-09-27T12:03:57.749946228Z"
//	}

//	func ApplyConfig(store interfaces.IStore, assets fs.FS) {
//		clients, err := store.GetClients(false)
//		if err != nil {
//			log.Error("Cannot get client config: ", err)
//			return
//		}
//		settings, err := store.GetGlobalSettings()
//		if err != nil {
//			log.Error("Cannot get global settings: ", err)
//			return
//		}
//		server, err := store.GetServer()
//		if err != nil {
//			log.Error("Cannot get server config: ", err)
//			return
//		}
//		utils.WriteWireGuardServerConfig(assets, server, clients, settings)
//	}
func GenNewClients(store interfaces.IStore) {
	for i := 0; i < 250; i++ {
		client := model.Client{
			Name:            fmt.Sprintf("client-%v", i),
			Email:           "",
			AllocatedIPs:    []string{fmt.Sprintf("10.7.0.%v/32", i+2)},
			AllowedIPs:      []string{"0.0.0.0/0"},
			ExtraAllowedIPs: []string{},
			Endpoint:        "",
			AdditionalNotes: "",
			UseServerDNS:    true,
			Enabled:         true,
		}
		NewClient(store, client)
	}
}

// NewClient handler
func NewClient(db interfaces.IStore, client model.Client) {
	// read server information
	server, err := db.GetServer()
	if err != nil {
		log.Error("Cannot fetch server from database: ", err)
		return
	}

	// validate the input Allocation IPs
	allocatedIPs, _ := utils.GetAllocatedIPs("")
	check, err := utils.ValidateIPAllocation(server.Interface.Addresses, allocatedIPs, client.AllocatedIPs)
	if !check {
		log.Error("Invalid Allocated IPs input from user: ", err)
		return
	}

	// validate the input AllowedIPs
	if utils.ValidateAllowedIPs(client.AllowedIPs) == false {
		log.Error("Allowed IPs must be in CIDR format")
		return
	}

	// validate extra AllowedIPs
	if utils.ValidateExtraAllowedIPs(client.ExtraAllowedIPs) == false {
		log.Warnf("Invalid Extra AllowedIPs input from user: %v", client.ExtraAllowedIPs)
		return
	}

	// gen ID
	guid := xid.New()
	client.ID = guid.String()

	// gen Wireguard key pair
	if client.PublicKey == "" {
		key, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			log.Error("Cannot generate wireguard key pair: ", err)
			return
		}
		client.PrivateKey = key.String()
		client.PublicKey = key.PublicKey().String()
	} else {
		_, err := wgtypes.ParseKey(client.PublicKey)
		if err != nil {
			log.Error("Cannot verify wireguard public key: ", err)
			return
		}
		// check for duplicates
		clients, err := db.GetClients(false)
		if err != nil {
			log.Error("Cannot get clients for duplicate check")
			return
		}
		for _, other := range clients {
			if other.Client.PublicKey == client.PublicKey {
				log.Error("Duplicate Public Key")
				return
			}
		}
	}

	if client.PresharedKey == "" {
		presharedKey, err := wgtypes.GenerateKey()
		if err != nil {
			log.Error("Cannot generated preshared key: ", err)
			return
		}
		client.PresharedKey = presharedKey.String()
	} else if client.PresharedKey == "-" {
		client.PresharedKey = ""
		log.Infof("skipped PresharedKey generation for user: %v", client.Name)
	} else {
		_, err := wgtypes.ParseKey(client.PresharedKey)
		if err != nil {
			log.Error("Cannot verify wireguard preshared key: ", err)
			return
		}
	}
	client.CreatedAt = time.Now().UTC()
	client.UpdatedAt = client.CreatedAt

	// write client to the database
	if err := db.SaveClient(client); err != nil {
		log.Error("Cannot save client to database: ", err)
		return
	}
	log.Infof("Created wireguard client: %v", client)

}

func initServerConfig(db interfaces.IStore, assetsDir fs.FS) bool {
	settings, err := db.GetGlobalSettings()
	if err != nil {
		log.Fatalf("Cannot get global settings: %v", err)
	}
	cients, _ := db.GetClients(false)
	if _, err := os.Stat(settings.ConfigFilePath); err == nil && len(cients) > 0 {
		// file exists, don't overwrite it implicitly
		return true
	}
	GenNewClients(db)
	// ApplyConfig(db, assetsDir)
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
	return false
}
