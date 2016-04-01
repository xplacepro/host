package lvm2

import (
	"log"
	"testing"
)

const (
	TEST_VG                 = "lxc"
	TEST_NON_EXISTENT_GROUP = "BPdORSznQB"
	TEST_LV                 = "zmdqoHZoN9"
	TEST_LV_SIZE_BYTES      = 5368709120 // 5g
)

func TestVgExists(t *testing.T) {
	vg := VolumeGroup{Name: TEST_VG}
	if err := vg.Exists(); err != nil {
		t.Error("Expected vg exists nil error")
	}
	vg_non_existent := VolumeGroup{Name: TEST_NON_EXISTENT_GROUP}
	if err := vg_non_existent.Exists(); err != nil {
		t.Error("Expected vg exists error")
	}
}

func TestVgInfo(t *testing.T) {
	vg := VolumeGroup{Name: TEST_VG}
	info, err := vg.Info()
	if err != nil {
		t.Error("Expected vg lxc exists nil error")
	}
	log.Println(info)
}

func TestCreateLv(t *testing.T) {
	vg := VolumeGroup{Name: TEST_VG}
	lv, create_err := vg.CreateLogicalVolume(TEST_LV, TEST_LV_SIZE_BYTES)
	if create_err != nil {
		t.Error("lv create error")
	}

	if err := lv.Exists(); err != nil {
		t.Error("Lv was not created")
	}
}

func TestLvInfo(t *testing.T) {
	vg := VolumeGroup{Name: TEST_VG}
	lv := vg.GetLv(TEST_LV)
	info, err := lv.Info()
	if err != nil {
		t.Error("Expected lv exists nil error")
	}
	log.Println(info)
}

func TestLvSize(t *testing.T) {
	vg := VolumeGroup{Name: TEST_VG}
	lv := vg.GetLv(TEST_LV)
	size, err := lv.Size()
	if err != nil {
		t.Error("Expected lv size nil error")
	}
	if size != TEST_LV_SIZE_BYTES {
		t.Error("Size not equal to creation param")
	}
}

func TestRemoveLv(t *testing.T) {
	vg := VolumeGroup{Name: TEST_VG}
	lv := vg.GetLv(TEST_LV)
	if _, err := lv.Remove(); err != nil {
		t.Error("Lv remove error")
	}
	if err := lv.Exists(); err != nil {
		t.Error("Lv exists after removing")
	}
}

func TestListLogicalVolumes(t *testing.T) {
	vg := VolumeGroup{Name: TEST_VG}
	lvs, err := vg.ListLogicalVolumes()
	if err != nil {
		t.Error("Error listing lvs, %s", err)
	}
	log.Println(lvs)
}
