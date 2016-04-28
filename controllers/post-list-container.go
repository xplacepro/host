package controllers

import (
	"encoding/json"
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type createContainerParams struct {
	Hostname string
	Dist     string
	Fssize   int
	User     string
	Password string
	Config   string
	Clone    string
}

func ValidatePostListContainer(c createContainerParams) bool {
	if strings.Trim(c.Dist, " ") == "" {
		return false
	}
	if strings.Trim(c.User, " ") == "" {
		return false
	}
	if strings.Trim(c.Password, " ") == "" {
		return false
	}
	if strings.Trim(c.Hostname, " ") == "" {
		return false
	}
	if c.Fssize <= 0 {
		return false
	}
	if len(c.Config) == 0 {
		return false
	}
	return true
}

func CreateContainer(lxc_c lxc.Container, create_params createContainerParams, config map[string]string) (interface{}, error) {
	meta := map[string]interface{}{}

	vgname, _ := config["lvm.lxc_vg"]
	out, err := lxc_c.Create(create_params.Dist, create_params.Fssize, vgname, create_params.Config, create_params.User, create_params.Password)
	if err != nil {
		return "", err
	}
	meta["output"] = out
	time.Sleep(2)

	conf, _ := lxc_c.ReadConfig()
	meta["config"] = conf
	ip_address, ip_err := lxc_c.GetInternalIp(30, true)
	if ip_err == nil {
		meta["internal_ipv4"] = ip_address
	}

	return meta, nil
}

func PostListContainerHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) rpc.Response {
	var create_params createContainerParams

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return rpc.BadRequest(err)
	}

	err = json.Unmarshal(data, &create_params)
	if err != nil {
		return rpc.BadRequest(err)
	}

	if !ValidatePostListContainer(create_params) {
		return rpc.BadRequest(ValidationError)
	}

	container := lxc.NewContainer(create_params.Hostname)
	if exists := container.Exists(); exists {
		return rpc.BadRequest(AlreadyExistsError)
	}

	crct := func(op *rpc.Operation) (interface{}, error) {
		return CreateContainer(*container, create_params, env.Config)
	}

	op_id := rpc.OperationCreate(crct, CREATE_OP_TYPE)

	return rpc.AsyncResponse(nil, op_id)

}
