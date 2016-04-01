package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"io/ioutil"
	"net/http"
)

type updateContainer struct {
	Config string
}

func PostContainerHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) rpc.Response {
	vars := mux.Vars(r)
	hostname := vars["hostname"]

	var c updateContainer

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rpc.BadRequest(err)
	}

	err = json.Unmarshal(data, &c)
	if err != nil {
		rpc.BadRequest(err)
	}

	container := lxc.NewContainer(hostname)

	if !container.Exists() {
		return rpc.NotFound
	}

	if err := container.ReplaceConfig(c.Config); err != nil {
		return rpc.InternalError(err)
	}

	return rpc.SyncResponse(nil)

}
