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
	"strings"
)

type resetPasswordParams struct {
	User     string
	Password string
}

func ValidateResetPassword(c resetPasswordParams) map[string]interface{} {
	validationErrors := make(map[string]interface{})

	if strings.Trim(c.User, " ") == "" {
		validationErrors["user"] = "user is required"
	}
	if strings.Trim(c.Password, " ") == "" {
		validationErrors["password"] = "password is required"
	}

	return validationErrors
}

func PostResetPasswordHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) (rpc.Response, int, error) {
	vars := mux.Vars(r)
	hostname := vars["hostname"]

	var resetParams resetPasswordParams

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: err}
	}

	err = json.Unmarshal(data, &resetParams)
	if err != nil {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: err}
	}

	validation_errors := ValidateResetPassword(resetParams)
	if len(validation_errors) > 0 {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: errors.New("Validation error"), MetadataMap: validation_errors}
	}

	container := lxc.NewContainer(hostname)

	if !container.Exists() {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: errors.New(fmt.Sprintf("%s doesn't exist", hostname))}
	}

	if err := container.ResetPassword(resetParams.User, resetParams.Password); err != nil {
		return nil, http.StatusBadRequest, rpc.StatusError{Err: err}
	}

	return rpc.SyncResponse{"Success", http.StatusOK, map[string]interface{}{}}, http.StatusOK, nil

}
