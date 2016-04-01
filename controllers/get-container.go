package controllers

import (
	"github.com/gorilla/mux"
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"net/http"
)

func GetContainerHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) rpc.Response {
	vars := mux.Vars(r)
	hostname := vars["hostname"]

	container := lxc.NewContainer(hostname)

	if !container.Exists() {
		return rpc.NotFound
	}

	info, err := container.Info()
	state := info["State"]

	if err != nil {
		return rpc.InternalError(err)
	}

	info_generic := make(map[string]interface{})

	for k, v := range info {
		info_generic[k] = v
	}

	if state == "RUNNING" {
		info_generic["Resources"] = container.Resources()
	}

	return rpc.SyncResponse(info_generic)
}
