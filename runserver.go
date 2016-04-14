package main

import (
	"flag"
	"github.com/gorilla/mux"
	"github.com/xplacepro/host/controllers"
	"github.com/xplacepro/rpc"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"syscall"
)

func ReloadEnv(env *rpc.Env, config map[string]string) {
	env.Auth = rpc.BasicAuthorization{config["auth.user"], config["auth.password"]}
	env.ClientAuth = rpc.ClientBasicAuthorization{config["client_auth.user"], config["client_auth.password"]}
	env.Config = config
}

func main() {

	if user, err := user.Current(); true {
		if err != nil {
			log.Fatal(err)
		}
		if user.Uid != "0" {
			log.Fatal("User must be root")
		}
	}

	var ConfigPath = flag.String("config", "config.ini", "Path to configuration file")

	flag.Parse()

	var config map[string]string

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	env := &rpc.Env{}

	go func() {
		for sign := range c {
			rpc.ParseConfiguration(*ConfigPath, &config)
			log.Printf("Reloading configuration, %s\n", sign)
			ReloadEnv(env, config)
		}
	}()

	rpc.ParseConfiguration(*ConfigPath, &config)

	ReloadEnv(env, config)

	r := mux.NewRouter()
	r.StrictSlash(false)
	r.Handle("/api/operations/1.0/{uuid:[a-zA-Z0-9-]+}/", rpc.Handler{env, rpc.GetOperationHandler}).Methods("GET")

	r.Handle("/api/containers/1.0/", rpc.Handler{env, controllers.GetListContainerHandler}).Methods("GET")
	r.Handle("/api/containers/1.0/", rpc.Handler{env, controllers.PostListContainerHandler}).Methods("POST")
	r.Handle("/api/containers/1.0/{hostname:[a-zA-Z0-9-_]+}/", rpc.Handler{env, controllers.GetContainerHandler}).Methods("GET")
	r.Handle("/api/containers/1.0/{hostname:[a-zA-Z0-9-_]+}/", rpc.Handler{env, controllers.PostContainerHandler}).Methods("POST")
	r.Handle("/api/containers/1.0/{hostname:[a-zA-Z0-9-_]+}/", rpc.Handler{env, controllers.DeleteContainerHandler}).Methods("DELETE")
	r.Handle("/api/containers/1.0/{hostname:[a-zA-Z0-9-_]+}/start/", rpc.Handler{env, controllers.PostStartContainerHandler}).Methods("POST")
	r.Handle("/api/containers/1.0/{hostname:[a-zA-Z0-9-_]+}/stop/", rpc.Handler{env, controllers.PostStopContainerHandler}).Methods("POST")
	r.Handle("/api/containers/1.0/{hostname:[a-zA-Z0-9-_]+}/reset-password/", rpc.Handler{env, controllers.PostResetPasswordHandler}).Methods("POST")
	r.Handle("/api/containers/1.0/{hostname:[a-zA-Z0-9-_]+}/backup/", rpc.Handler{env, controllers.PostBackupContainerHandler}).Methods("POST")
	r.Handle("/api/containers/1.0/{hostname:[a-zA-Z0-9-_]+}/restore/", rpc.Handler{env, controllers.PostRestoreContainerHandler}).Methods("POST")
	http.Handle("/", r)
	log.Printf("Started server on %s", config["listen"])
	panic(http.ListenAndServe(config["listen"], nil))
}
