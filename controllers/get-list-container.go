package controllers

import (
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"net/http"
)

func GetListContainerHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) rpc.Response {
	containers, err := lxc.ListContainers()
	response := make(map[string]interface{})

	if err != nil {
		return rpc.InternalError(err)
	}

	response["Сontainers"] = containers
	return rpc.SyncResponse(response)
}
