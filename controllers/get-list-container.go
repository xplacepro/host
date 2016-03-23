package controllers

import (
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"net/http"
)

func GetListContainerHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) (rpc.Response, int, error) {
	containers, err := lxc.ListContainers()
	response := make(map[string]interface{})

	if err != nil {
		return nil, http.StatusInternalServerError, rpc.StatusError{Err: err}
	}

	response["Сontainers"] = containers
	return rpc.SyncResponse{"Success", http.StatusOK, response}, http.StatusOK, nil
}
