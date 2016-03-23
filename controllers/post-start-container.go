package controllers

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"net/http"
)

func PostStartContainerHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) (rpc.Response, int, error) {
	vars := mux.Vars(r)
	hostname := vars["hostname"]

	container := lxc.NewContainer(hostname)

	if !container.Exists() {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: errors.New(fmt.Sprintf("%s doesn't exist", hostname))}
	}

	if err := container.Start(); err != nil {
		return nil, http.StatusInternalServerError, rpc.StatusError{Err: err}
	}

	return rpc.SyncResponse{"Success", http.StatusOK, map[string]interface{}{}}, http.StatusOK, nil
}
