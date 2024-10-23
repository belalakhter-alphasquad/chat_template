package api

import (
	"fmt"
	"net/http"

	"github.com/belalakhter-alphasquad/chat_template/internal/chat"
	"github.com/belalakhter-alphasquad/chat_template/utils"
	"github.com/gorilla/mux"
)

type server interface {
}

type Server struct {
	Web *http.Server
}

func NewServer() *Server {
	r := SetupRouter()
	Addr := ":6003"

	srv := &http.Server{
		Addr:    Addr,
		Handler: r,
	}
	server := Server{
		Web: srv,
	}
	utils.LogMessage(fmt.Sprintf("Listening Server on %v ...", Addr), 3)
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			utils.LogMessage(err.Error(), 1)
		}
	}()

	return &server

}

func SetupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		chat.UpgradeConnectionWs(w, r)
	})
	return r
}
