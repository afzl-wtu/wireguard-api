package api

import "net/http"

type ApiServer struct {
	Mux  *http.ServeMux
	Addr string
}

func NewApiServer() *ApiServer {
	return &ApiServer{
		Mux:  http.NewServeMux(),
		Addr: ":8080",
	}
}

func (a *ApiServer) Start() error {
	a.globalHandlers()
	return http.ListenAndServe(a.Addr, a.Mux)
}

func (a *ApiServer) globalHandlers() {

	a.Mux.HandleFunc("/api/v1", a.statusHandler)

}

func (a *ApiServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<b>It is wireguard api</b><br> Hi"))
}
