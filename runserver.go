package main

import (
	"flag"
	"github.com/gorilla/mux"
	"github.com/xplacepro/host/controllers"
	"github.com/xplacepro/rpc"
	"log"
	"net/http"
)

func main() {
	var ConfigPath = flag.String("config", "config.ini", "Path to configuration file")
	flag.Parse()

	config, err := rpc.ParseConfiguration(*ConfigPath)

	if err != nil {
		panic(err)
	}

	env := &rpc.Env{Auth: rpc.BasicAuthorization{config["auth.user"], config["auth.password"]},
		ClientAuth: rpc.ClientBasicAuthorization{config["client_auth.user"], config["client_auth.password"]}}

	r := mux.NewRouter()
	r.StrictSlash(false)
	r.Handle("/api/v1/containers", rpc.Handler{env, controllers.GetListContainerHandler}).Methods("GET")
	r.Handle("/api/v1/containers", rpc.Handler{env, controllers.PostListContainerHandler}).Methods("POST")
	r.Handle("/api/v1/containers/{hostname:[a-zA-Z0-9-]+}", rpc.Handler{env, controllers.GetContainerHandler}).Methods("GET")
	r.Handle("/api/v1/containers/{hostname:[a-zA-Z0-9-]+}", rpc.Handler{env, controllers.PostContainerHandler}).Methods("POST")
	r.Handle("/api/v1/containers/{hostname:[a-zA-Z0-9-]+}", rpc.Handler{env, controllers.DeleteContainerHandler}).Methods("DELETE")
	r.Handle("/api/v1/containers/{hostname:[a-zA-Z0-9-]+}/start", rpc.Handler{env, controllers.PostStartContainerHandler}).Methods("POST")
	r.Handle("/api/v1/containers/{hostname:[a-zA-Z0-9-]+}/stop", rpc.Handler{env, controllers.PostStopContainerHandler}).Methods("POST")
	r.Handle("/api/v1/containers/{hostname:[a-zA-Z0-9-]+}/reset-password", rpc.Handler{env, controllers.PostResetPasswordHandler}).Methods("POST")
	http.Handle("/", r)
	log.Printf("Started server on %s", config["listen"])
	panic(http.ListenAndServe(config["listen"], nil))
}
