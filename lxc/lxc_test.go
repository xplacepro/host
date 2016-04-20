package lxc

import (
	"fmt"
	"log"
	"regexp"
	"testing"
)

func TestGenerateNewMacAddr(t *testing.T) {
	mac, err := GenerateNewMacAddr()
	if err != nil {
		t.Error("Expected mac addr")
	}
	log.Println(mac)
}

func TestReplaceMacAddr(t *testing.T) {
	orig := "lxc.network.hwaddr = 00:16:3e:3a:56:e1"
	re := regexp.MustCompile(`lxc.network.hwaddr\s*=\s*[0-9a-zA-Z:]+`)
	macAddr, _ := GenerateNewMacAddr()
	conf := re.ReplaceAllString(orig, fmt.Sprintf("%s = %s", "lxc.network.hwaddr", macAddr))
	log.Printf(conf)
}
