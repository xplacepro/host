package controllers

import (
	"github.com/gorilla/mux"
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"net/http"
)

func PostStopContainerHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) rpc.Response {
	vars := mux.Vars(r)
	hostname := vars["hostname"]

	container := lxc.NewContainer(hostname)

	if !container.Exists() {
		return rpc.NotFound
	}

	if err := container.Stop(); err != nil {
		return rpc.InternalError(err)
	}

	return rpc.SyncResponse(nil)
}
