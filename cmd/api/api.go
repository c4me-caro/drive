package api

import (
	"net/http"

	"github.com/c4me-caro/drive/cmd/auth"
	"github.com/c4me-caro/drive/database"
	"github.com/c4me-caro/drive/service/driver"
	"github.com/c4me-caro/drive/service/user"
	"github.com/gorilla/mux"
)

type APIServer struct {
	addr string
	db   *database.DriveWorker
}

func NewApiServer(addr string, db *database.DriveWorker) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter().StrictSlash(true)
	subrouter := router.PathPrefix("/drive").Subrouter()

	userHandler := user.NewHandler(s.db)
	userHandler.RegisterRoutes(router)

	driverHandler := driver.NewHandler(s.db)
	driverHandler.RegisterRoutes(subrouter)

	router.Use(auth.HandleAuthorization)

	service := &http.Server{
		Handler: router,
		Addr:    s.addr,
	}

	return service.ListenAndServe()
}
