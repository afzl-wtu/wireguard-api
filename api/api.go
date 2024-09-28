package api

import (
	"net/http"

	"github.com/afzl-wtu/wireguard-api/handler"
	"github.com/afzl-wtu/wireguard-api/interfaces"
)

type ApiServer struct {
	Mux  *http.ServeMux
	Addr string
}

func NewApiServer() *ApiServer {
	return &ApiServer{
		Mux:  http.NewServeMux(),
		Addr: ":2200",
	}
}

func (a *ApiServer) Start(store interfaces.IStore) error {
	a.globalHandlers(store)
	return http.ListenAndServe(a.Addr, a.Mux)
}

func (a *ApiServer) globalHandlers(store interfaces.IStore) {
	a.Mux.HandleFunc("GET /api/v1/getconfig", handler.GetConfig(store))
}
