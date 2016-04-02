package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/xplacepro/common"
	"github.com/xplacepro/host/lvm2"
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"io/ioutil"
	"log"
	"net/http"
)

type backupContainer struct {
	Sync bool
}

func GoBackupContainer(lxc_c lxc.Container, config map[string]string) (interface{}, error) {
	log.Printf("Backing up container: %v", lxc_c)
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
		size, _ := lv.Size()
		if _, err := backupVg.CreateLogicalVolume(BackupName(lxc_c.Name), size); err != nil {
			return nil, err
		}
	}
	meta := make(map[string]interface{})
	out, err := common.Dd(lv.FullName(), backupLv.FullName())
	if err != nil {
		return nil, err
	}
	meta["output"] = out

	log.Printf("Backed up container: %v, result: %v, err: %v", lxc_c, out, err)

	return meta, nil
}

func PostBackupContainerHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) rpc.Response {
	vars := mux.Vars(r)
	hostname := vars["hostname"]

	var params backupContainer

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rpc.BadRequest(err)
	}

	err = json.Unmarshal(data, &params)
	if err != nil {
		rpc.BadRequest(err)
	}

	container := lxc.NewContainer(hostname)

	if !container.Exists() {
		return rpc.NotFound
	}

	if state, _ := container.GetState(); state != "STOPPED" {
		return rpc.BadRequest(NotStopped(hostname))
	}

	if params.Sync {
		GoBackupContainer(*container, env.Config)
		return rpc.SyncResponse(nil)
	} else {
		dlct := func(op *rpc.Operation) (interface{}, error) {
			return GoBackupContainer(*container, env.Config)
		}

		op_id, _ := rpc.OperationCreate(dlct, BACKUP_OP_TYPE)

		return rpc.AsyncResponse(nil, op_id)
	}

}
