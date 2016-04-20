package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/pborman/uuid"
	"github.com/xplacepro/common"
	"github.com/xplacepro/host/lvm2"
	"github.com/xplacepro/host/lxc"
	"github.com/xplacepro/rpc"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type cloneContainerParams struct {
	Original string
	Hostname string
}

func ValidatePostCloneContainer(c cloneContainerParams) bool {
	if strings.Trim(c.Original, " ") == "" {
		return false
	}
	if strings.Trim(c.Hostname, " ") == "" {
		return false
	}
	return true
}

func CloneContainer(originalContainer *lxc.Container, newContainer *lxc.Container, clone_params cloneContainerParams, config map[string]string) (interface{}, error) {
	log.Printf("Cloning container: %v, params: %v", clone_params.Original, clone_params)
	meta := map[string]interface{}{}

	if err := os.MkdirAll(newContainer.BasePath(), 0770); err != nil {
		return nil, err
	}

	macAddr, macErr := lxc.GenerateNewMacAddr()
	if macErr != nil {
		return nil, macErr
	}

	re := regexp.MustCompile(`lxc.network.hwaddr\s*=\s*[0-9a-zA-Z:]+`)
	originalConf, confErr := originalContainer.ReadConfig()
	if confErr != nil {
		return nil, confErr
	}

	conf := re.ReplaceAllString(originalConf, fmt.Sprintf("%s = 00:16:3e:%s", "lxc.network.hwaddr", macAddr))
	conf = strings.Replace(conf, originalContainer.Name, newContainer.Name, -1)

	if err := newContainer.ReplaceConfig(conf); err != nil {
		return nil, err
	}

	meta["config"] = conf

	vgname := config["lvm.lxc_vg"]

	vg := lvm2.VolumeGroup{Name: vgname}

	if err := vg.Exists(); err != nil {
		return nil, err
	}

	originalLv := vg.GetLv(clone_params.Original)
	if err := originalLv.Exists(); err != nil {
		return nil, err
	}
	originalSize, _ := originalLv.Size()

	cloneLv, create_err := vg.CreateLogicalVolume(clone_params.Hostname, originalSize)

	if create_err != nil {
		return nil, create_err
	}

	if _, err := common.Mkfs_ext4(cloneLv.FullName()); err != nil {
		return nil, err
	}

	tmpUuid := uuid.New()
	originalPath := fmt.Sprintf("/tmp/%s/%s", clone_params.Original, tmpUuid)
	clonePath := fmt.Sprintf("/tmp/%s/%s", clone_params.Hostname, tmpUuid)

	if err := os.MkdirAll(originalPath, 644); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(clonePath, 644); err != nil {
		return nil, err
	}

	if err := vg.Exists(); err != nil {
		return nil, err
	}

	if _, err := common.RunCommand("mount", []string{originalLv.FullName(), originalPath}); err != nil {
		return nil, err
	}

	if _, err := common.RunCommand("mount", []string{cloneLv.FullName(), clonePath}); err != nil {
		return nil, err
	}

	time.Sleep(1)

	cleanUp := func() {
		common.RunCommand("umount", []string{originalPath})
		common.RunCommand("umount", []string{clonePath})
	}

	if _, err := common.Rsync(fmt.Sprintf("%s/", originalPath), clonePath); err != nil {
		cleanUp()
		return nil, err
	}

	if err := common.ReplaceInFile(path.Join(clonePath, "/etc/hostname"), clone_params.Original, clone_params.Hostname, 0644); err != nil {
		cleanUp()
		return nil, err
	}

	if err := common.ReplaceInFile(path.Join(clonePath, "/etc/hosts"), clone_params.Original, clone_params.Hostname, 0644); err != nil {
		cleanUp()
		return nil, err
	}

	log.Printf("Cloned container: %v", clone_params)

	cleanUp()

	defer originalContainer.Stop()
	if ip_address, ip_err := originalContainer.GetInternalIp(30, false); err != nil {
		return nil, ip_err
	}

	time.Sleep(2)

	ip_address, ip_err := newContainer.GetInternalIp(30, true)
	if ip_err == nil {
		meta["internal_ipv4"] = ip_address
	}

	return meta, nil
}

func PostCloneContainerHandler(env *rpc.Env, w http.ResponseWriter, r *http.Request) rpc.Response {
	var clone_params cloneContainerParams

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return rpc.BadRequest(err)
	}

	err = json.Unmarshal(data, &clone_params)
	if err != nil {
		return rpc.BadRequest(err)
	}

	if !ValidatePostCloneContainer(clone_params) {
		return rpc.BadRequest(ValidationError)
	}

	container := lxc.NewContainer(clone_params.Hostname)
	if exists := container.Exists(); exists {
		return rpc.BadRequest(AlreadyExistsError)
	}

	originalContainer := lxc.NewContainer(clone_params.Original)
	if exists := originalContainer.Exists(); !exists {
		return rpc.BadRequest(DoesNotExistError)
	}

	if state, _ := originalContainer.GetState(); state != "STOPPED" {
		return rpc.BadRequest(NotStopped(originalContainer.Name))
	}

	crct := func(op *rpc.Operation) (interface{}, error) {
		return CloneContainer(originalContainer, container, clone_params, env.Config)
	}

	op_id, _ := rpc.OperationCreate(crct, CLONE_OP_TYPE)

	return rpc.AsyncResponse(nil, op_id)

}
