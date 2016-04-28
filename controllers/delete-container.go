package controllers

import (
	"github.com/gorilla/mux"
	"github.com/xplacepro/host/lvm2"
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"log"
	"net/http"
)

func DeleteContainer(lxc_c lxc.Container, config map[string]string) (interface{}, error) {
	log.Printf("Deleting container: %v", lxc_c)
	meta := map[string]interface{}{}
	out, err := lxc_c.Destroy()
	meta["output"] = out
	if err != nil {
		return "", err
	}

	backupVg := lvm2.VolumeGroup{Name: config["lvm.backup_vg"]}
	if err := backupVg.Exists(); err != nil {
		return nil, err
	}

	backupLv := backupVg.GetLv(BackupName(lxc_c.Name))

	if err := backupLv.Exists(); err == nil {
		r_out, r_err := backupLv.Remove()
		if r_err != nil {
			return nil, r_err
		}
		meta["lvremove"] = r_out
	}

	log.Printf("Deleted container: %v, result: %v, err: %v", lxc_c, out, err)
	return meta, nil
}

func DeleteContainerHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) rpc.Response {
	vars := mux.Vars(r)
	hostname := vars["hostname"]

	container := lxc.NewContainer(hostname)

	if !container.Exists() {
		return rpc.NotFound
	}

	if state, _ := container.GetState(); state != "STOPPED" {
		return rpc.BadRequest(NotStopped(hostname))
	}

	dlct := func(op *rpc.Operation) (interface{}, error) {
		return DeleteContainer(*container, env.Config)
	}

	op_id := rpc.OperationCreate(dlct, "DELETE_OP_TYPE")

	return rpc.AsyncResponse(nil, op_id)
}
