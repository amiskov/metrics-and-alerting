package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

type metricsAPI struct {
	Router *chi.Mux
	store  Storage
}

func New(s Storage) *metricsAPI {
	api := &metricsAPI{
		Router: chi.NewRouter(),
		store:  s,
	}
	api.mountHandlers()
	return api
}

func (api *metricsAPI) Run(address string) {
	fmt.Printf("Server has been started at http://%s\n", address)

	server := &http.Server{
		Addr:    address,
		Handler: api.Router,
	}

	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error listening for server: %s\n", err)
	}
}
