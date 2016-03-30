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
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
)

func ReloadEnv(env *rpc.Env, config map[string]string) {
	env.Auth = rpc.BasicAuthorization{config["auth.user"], config["auth.password"]}
	env.ClientAuth = rpc.ClientBasicAuthorization{config["client_auth.user"], config["client_auth.password"]}
}

func DoNotifyState(notifier, hostname, state string, env *rpc.Env) error {
	url := fmt.Sprintf(notifier, hostname, state)
	_, err := rpc.DoCallbackRequest(url, rpc.CallbackRequest{}, env.ClientAuth)
	return err
}

func main() {
	var ConfigPath = flag.String("config", "config.ini", "Path to configuration file")
	var NotifyState = flag.String("notify", "", "Notify container state")
	var Hostname = flag.String("hostname", "", "Notify container hostname")
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

	if *NotifyState != "" && *Hostname != "" {
		DoNotifyState(config["state.notifier"], *Hostname, *NotifyState, env)
		return
	}

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
