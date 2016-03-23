package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type DeleteContainerParams struct {
	Callback string
	Code     string
}

func ValidateDeleteContainer(c DeleteContainerParams) map[string]interface{} {
	validationErrors := make(map[string]interface{})

	if strings.Trim(c.Code, " ") == "" {
		validationErrors["code"] = "code is required"
	}
	if !IsURL(c.Callback) {
		validationErrors["callback"] = "valid url is required"
	}

	return validationErrors
}

func GoDeleteContainer(lxc_c lxc.Container, delete_params DeleteContainerParams, env *rpc.Env) {
	log.Printf("Deleting container: %v", lxc_c)
	out, err := lxc_c.Destroy()
	log.Printf("Deleted container: %v, result: %v, err: %v", lxc_c, out, err)
	callback_req := rpc.CallbackRequest{Err: err, Identifier: lxc_c.Name, Op_type: "DELETE_CONTAINER", Metadata: map[string]interface{}{}, Code: delete_params.Code, Output: out}
	log.Printf("Making callback request to %s: %v", delete_params.Callback, callback_req)
	rpc.DoCallbackRequest(delete_params.Callback, callback_req, env.ClientUser, env.ClientPassword)
}

func DeleteContainerHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) (rpc.Response, int, error) {
	vars := mux.Vars(r)
	hostname := vars["hostname"]

	container := lxc.NewContainer(hostname)

	var delete_params DeleteContainerParams

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: err}
	}

	err = json.Unmarshal(data, &delete_params)
	if err != nil {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: err}
	}

	validation_errors := ValidateDeleteContainer(delete_params)
	if len(validation_errors) > 0 {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: errors.New("Validation error"), MetadataMap: validation_errors}
	}

	if !container.Exists() {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: errors.New(fmt.Sprintf("%s doesn't exist", hostname))}
	}

	if state, _ := container.GetState(); state != "STOPPED" {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: errors.New(fmt.Sprintf("%s is not in STOPPED state", hostname))}
	}

	go GoDeleteContainer(*container, delete_params, env)

	return rpc.SyncResponse{"Success", http.StatusOK, map[string]interface{}{}}, http.StatusOK, nil
}
