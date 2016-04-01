package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"io/ioutil"
	"net/http"
	"strings"
)

type resetPasswordParams struct {
	User     string
	Password string
}

func ValidateResetPassword(c resetPasswordParams) bool {
	if strings.Trim(c.User, " ") == "" {
		return false
	}
	if strings.Trim(c.Password, " ") == "" {
		return false
	}

	return true
}

func PostResetPasswordHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) rpc.Response {
	vars := mux.Vars(r)
	hostname := vars["hostname"]

	var resetParams resetPasswordParams

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return rpc.BadRequest(err)
	}

	err = json.Unmarshal(data, &resetParams)
	if err != nil {
		return rpc.BadRequest(err)
	}

	if !ValidateResetPassword(resetParams) {
		return rpc.BadRequest(ValidationError)
	}

	container := lxc.NewContainer(hostname)

	if !container.Exists() {
		return rpc.NotFound
	}

	if err := container.ResetPassword(resetParams.User, resetParams.Password); err != nil {
		return rpc.InternalError(err)
	}

	return rpc.SyncResponse(nil)

}
