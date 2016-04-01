package lvm2

import (
	"errors"
	"fmt"
	"github.com/xplacepro/common"
	"path"
	"strconv"
	"strings"
)

type VolumeGroup struct {
	Name string
}

type LogicalVolume struct {
	Name string
	Vg   *VolumeGroup
}

func (vg *VolumeGroup) FullName() string {
	return path.Join("/dev", vg.Name)
}

func (vg *VolumeGroup) Exists() error {
	if _, err := common.RunCommand("vgs", []string{vg.Name}); err != nil {
		return err
	}
	return nil
}

func (vg *VolumeGroup) Info() (map[string]string, error) {
	out, err := common.RunCommand("vgs", []string{vg.Name, "--rows", "--unit=b", "--separator=:"})
	if err != nil {
		return nil, err
	}
	values := common.ParseValues(out, rune(':'), '#')
	return values, nil
}

func (vg *VolumeGroup) CreateLogicalVolume(name string, fssize int) (LogicalVolume, error) {
	_, err := common.RunCommand("lvcreate", []string{"-L", fmt.Sprintf("%vb", fssize), "-n", name, vg.Name})
	return LogicalVolume{Name: name, Vg: vg}, err
}

func (vg *VolumeGroup) ListLogicalVolumes() ([]LogicalVolume, error) {
	out, err := common.RunCommand("lvs", []string{vg.FullName(), "--unit=b", "-o", "lv_name", "--noheading"})
	if err != nil {
		return nil, err
	}
	items := strings.Fields(out)
	lvs := make([]LogicalVolume, len(items))
	for idx, name := range items {
		lvs[idx] = LogicalVolume{Name: name, Vg: vg}
	}
	return lvs, nil
}

func (vg *VolumeGroup) GetLv(name string) LogicalVolume {
	return LogicalVolume{Name: name, Vg: vg}
}

func (lv *LogicalVolume) FullName() string {
	return path.Join(lv.Vg.FullName(), lv.Name)
}

func (lv *LogicalVolume) Remove() (string, error) {
	return common.RunCommand("lvremove", []string{lv.FullName(), "-f"})
}

func (lv *LogicalVolume) Exists() error {
	if _, err := common.RunCommand("lvs", []string{lv.FullName()}); err != nil {
		return err
	}
	return nil
}

func (lv *LogicalVolume) Info() (map[string]string, error) {
	out, err := common.RunCommand("lvs", []string{lv.FullName(), "--rows", "--unit=b", "--separator=:"})
	if err != nil {
		return nil, err
	}
	values := common.ParseValues(out, rune(':'), '#')
	return values, nil
}

func (lv *LogicalVolume) Size() (int, error) {
	info, err := lv.Info()
	if err != nil {
		return 0, err
	}
	size, ok := info["LSize"]
	if !ok {
		return 0, errors.New("Size not found")
	}
	size = strings.TrimRight(size, "B")
	size = strings.TrimRight(size, "b")
	return strconv.Atoi(size)
}
