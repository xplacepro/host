package controllers

import (
	"encoding/json"
	"errors"
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type CreateContainerParams struct {
	Hostname string
	Dist     string
	Fssize   int
	User     string
	Password string
	Config   string
	Callback string
	Code     string
}

func ValidatePostListContainer(c CreateContainerParams) map[string]interface{} {
	validationErrors := make(map[string]interface{})

	if strings.Trim(c.Dist, " ") == "" {
		validationErrors["dist"] = "dist is required"
	}
	if strings.Trim(c.User, " ") == "" {
		validationErrors["user"] = "user is required"
	}
	if strings.Trim(c.Password, " ") == "" {
		validationErrors["password"] = "password is required"
	}
	if strings.Trim(c.Hostname, " ") == "" {
		validationErrors["hostname"] = "hostname is required"
	}
	if c.Fssize <= 0 {
		validationErrors["fssize"] = "fssize more than zero is required"
	}
	if len(c.Config) == 0 {
		validationErrors["config"] = "config is required"
	}
	if strings.Trim(c.Code, " ") == "" {
		validationErrors["code"] = "code is required"
	}
	if !IsURL(c.Callback) {
		validationErrors["callback"] = "valid url is required"
	}

	return validationErrors
}

func GoCreateContainer(lxc_c lxc.Container, create_params CreateContainerParams, env *rpc.Env) {
	log.Printf("Creating container: %v, params: %v", lxc_c, create_params)
	out, err := lxc_c.Create(create_params.Dist, create_params.Fssize, create_params.Config)
	log.Printf("Created container: %v, params: %v, result: %v, err: %v", lxc_c, create_params, out, err)
	conf, _ := lxc_c.ReadConfig()
	ip_address := lxc_c.GetInternalIp()
	meta := map[string]interface{}{"config": conf, "internal_ipv4": ip_address}
	callback_req := rpc.CallbackRequest{Err: err, Identifier: lxc_c.Name, Op_type: "CREATE_CONTAINER", Metadata: meta, Code: create_params.Code, Output: out}
	log.Printf("Making callback request to %s: %v", create_params.Callback, callback_req)
	rpc.DoCallbackRequest(create_params.Callback, callback_req, env.ClientUser, env.ClientPassword)
}

func PostListContainerHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) (rpc.Response, int, error) {
	var create_params CreateContainerParams

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: err}
	}

	err = json.Unmarshal(data, &create_params)
	if err != nil {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: err}
	}

	validation_errors := ValidatePostListContainer(create_params)
	if len(validation_errors) > 0 {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: errors.New("Validation error"), MetadataMap: validation_errors}
	}

	container := lxc.NewContainer(create_params.Hostname)
	if exists := container.Exists(); exists {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: errors.New("Container already exists")}
	}

	go GoCreateContainer(*container, create_params, env)

	return rpc.AsyncResponse{"OK", http.StatusOK, map[string]interface{}{}}, http.StatusOK, nil

}
