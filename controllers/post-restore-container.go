package controllers

import (
	"github.com/gorilla/mux"
	"github.com/xplacepro/common"
	"github.com/xplacepro/host/lvm2"
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"log"
	"net/http"
)

func RestoreContainer(lxc_c lxc.Container, config map[string]string) (interface{}, error) {
	lxcVg := lvm2.VolumeGroup{Name: config["lvm.lxc_vg"]}

	if err := lxcVg.Exists(); err != nil {
		return nil, err
	}

	lv := lxcVg.GetLv(lxc_c.Name)

	if err := lv.Exists(); err != nil {
		return nil, err
	}

	backupVg := lvm2.VolumeGroup{Name: config["lvm.backup_vg"]}
	if err := backupVg.Exists(); err != nil {
		return nil, err
	}

	backupLv := backupVg.GetLv(BackupName(lxc_c.Name))

	if err := backupLv.Exists(); err != nil {
		return nil, err
	}
	meta := make(map[string]interface{})
	out, err := common.Dd(backupLv.FullName(), lv.FullName())
	if err != nil {
		return nil, err
	}
	meta["output"] = out

	return meta, nil
}

func PostRestoreContainerHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) rpc.Response {
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
		return RestoreContainer(*container, env.Config)
	}

	op_id := rpc.OperationCreate(dlct, RESTORE_OP_TYPE)

	return rpc.AsyncResponse(nil, op_id)
}
