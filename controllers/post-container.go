package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"io/ioutil"
	"net/http"
)

type updateContainer struct {
	Config string
}

func PostContainerHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) (rpc.Response, int, error) {
	vars := mux.Vars(r)
	hostname := vars["hostname"]

	var c updateContainer

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: err}
	}

	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: err}
	}

	container := lxc.NewContainer(hostname)

	if !container.Exists() {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: errors.New(fmt.Sprintf("%s doesn't exist", hostname))}
	}

	if err := container.ReplaceConfig(c.Config); err != nil {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: err}
	}

	return rpc.SyncResponse{"Success", http.StatusOK, map[string]interface{}{}}, http.StatusOK, nil

}
