package handler

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/afzl-wtu/wireguard-api/interfaces"
	model "github.com/afzl-wtu/wireguard-api/models"
	"github.com/afzl-wtu/wireguard-api/utils"
)

type abc func(http.ResponseWriter, *http.Request)

var mutex = sync.Mutex{}

func GetConfig(store interfaces.IStore) abc {
	clients, _ := store.GetClients(false)
	clientPublicKeys := make(map[string]*model.Client)
	for _, clientData := range clients {
		clientPublicKeys[clientData.Client.PublicKey] = clientData.Client
	}
	return func(w http.ResponseWriter, r *http.Request) {
		uid := r.URL.Query().Get("uid")
		if uid == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Please provide a valid client ID"))
			return
		}

		mutex.Lock()
		inActiveClients := utils.GetInActiveClients()
		mutex.Unlock()
		if inActiveClients == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to get active clients"))
			return
		}
		selectedClient := clientPublicKeys[inActiveClients[0]]
		json.NewEncoder(w).Encode(selectedClient)

	}
}
