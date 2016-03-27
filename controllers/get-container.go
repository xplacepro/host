package controllers

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"net/http"
)

func GetContainerHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) (rpc.Response, int, error) {
	vars := mux.Vars(r)
	hostname := vars["hostname"]

	container := lxc.NewContainer(hostname)

	if !container.Exists() {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: errors.New(fmt.Sprintf("%s doesn't exist", hostname))}
	}

	info, err := container.Info()
	state := info["State"]

	if err != nil {
		return nil, http.StatusInternalServerError, rpc.StatusError{Err: err}
	}

	info_generic := make(map[string]interface{})

	for k, v := range info {
		info_generic[k] = v
	}

	if state == "RUNNING" {
		info_generic["Resources"] = container.Resources()
	}

	return rpc.SyncResponse{"Success", http.StatusOK, info_generic}, http.StatusOK, nil
}
