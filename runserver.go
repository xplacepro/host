package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/xplacepro/host/controllers"
	"github.com/xplacepro/rpc"
	"log"
	"net/http"
	"os"
)

func Usage() {
	fmt.Println("Usage:")
	flag.PrintDefaults()
	os.Exit(0)
}

func main() {
	var User = flag.String("user", "", "Basic auth user")
	var Password = flag.String("password", "", "Basic auth password")

	var ClientUser = flag.String("client-user", "", "Basic auth user for dashboard client")
	var ClientPassword = flag.String("client-password", "", "Basic auth user password for dashboard client")

	var Listen = flag.String("listen", ":8080", "Interface and port to listen, :8080")

	flag.Parse()

	if *User == "" || *Password == "" {
		Usage()
	}

	env := &rpc.Env{Auth: rpc.BasicAuthorization{*User, *Password}, ClientUser: *ClientUser, ClientPassword: *ClientPassword}
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
	log.Printf("Started server on %s", *Listen)
	http.ListenAndServe(*Listen, nil)
}
