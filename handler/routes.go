package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/afzl-wtu/wireguard-api/interfaces"
	model "github.com/afzl-wtu/wireguard-api/models"
	"github.com/afzl-wtu/wireguard-api/utils"
	log "github.com/sirupsen/logrus"
)

type abc func(http.ResponseWriter, *http.Request)

var mutex = sync.Mutex{}

type ReservedClient struct {
	ClientID string
	Time     time.Time
}

func GetConfig(store interfaces.IStore) abc {
	reserverdConfigs := []ReservedClient{}
	clients, _ := store.GetClients(false)
	server, _ := store.GetServer()
	globalSettings, _ := store.GetGlobalSettings()
	clientPublicKeys := make(map[string]*model.Client)
	for _, clientData := range clients {
		clientPublicKeys[clientData.Client.PublicKey] = clientData.Client
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if len(clients) == 0 {
			w.Write([]byte("No clients found"))
		}
		uid := r.URL.Query().Get("uid")
		if uid == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Please provide a valid client ID"))
			return
		}

		mutex.Lock()
		inActiveClients := utils.GetInActiveClients()
		if inActiveClients == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to get inactive clients"))
			return
		}
		var clientToSend model.Client
		for _, inActiveClientString := range inActiveClients {
			selectedInActiveClient := clientPublicKeys[inActiveClientString]
			// check if slectedClient is reserved
			index := -1
			for i, reservedClient := range reserverdConfigs {
				if reservedClient.ClientID == selectedInActiveClient.ID {
					index = i
					break
				}
			}
			if index == -1 {
				clientToSend = *selectedInActiveClient
				reserverdConfigs = append(reserverdConfigs, ReservedClient{ClientID: selectedInActiveClient.ID, Time: time.Now()})
				break
			} else {
				if time.Since(reserverdConfigs[index].Time) > time.Minute*1 {
					clientToSend = *selectedInActiveClient
					reserverdConfigs[index].Time = time.Now()
					break
				}
			}

		}
		mutex.Unlock()
		download := utils.BuildClientConfig(clientToSend, server, globalSettings)
		log.Info("Total Reserved Configs: ", len(reserverdConfigs))
		w.Write([]byte(download))
	}
}
