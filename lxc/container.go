package lxc

import (
	"bytes"
	"fmt"
	"github.com/xplacepro/common"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	LXC_BASE_PATH = "/var/lib/lxc"
	LXC_BIN       = "/usr/bin"
)

type Container struct {
	Name   string
	State  string
	Config string
}

func (c *Container) GetState() (string, error) {
	if c.State != "" {
		return c.State, nil
	}
	info, err := c.Info()
	if err != nil {
		return "", err
	}
	state, _ := info["State"]
	return state, nil

}

func (c *Container) Start() error {
	_, err := common.RunCommand(path.Join(LXC_BIN, "lxc-start"), []string{"-n", c.Name})
	return err
}

func (c *Container) Stop() error {
	_, err := common.RunCommand(path.Join(LXC_BIN, "lxc-stop"), []string{"-n", c.Name})
	if err != nil && !strings.Contains(err.Error(), "is not running") {
		return err
	}
	return nil
}

func (c *Container) Info() (map[string]string, error) {
	out, err := common.RunCommand(path.Join(LXC_BIN, "lxc-info"), []string{"-n", c.Name})
	if err != nil {
		return nil, err
	}
	values := common.ParseValues(out, rune(':'), '#')
	return values, nil
}

func (c *Container) Resources() map[string]interface{} {
	resources := make(map[string]interface{})
	if cpu, err := c.CpuUsage(); err == nil {
		resources["Cpu"] = cpu
	}
	if ram, err := c.RamUsage(); err == nil {
		resources["Ram"] = ram
	}
	resources["Timestamp"] = time.Now().Unix()
	return resources
}

func (c *Container) GetInternalIp(timeout int) (string, error) {
	if err := c.Start(); err != nil {
		return "", err
	}

	ip_chan := make(chan string)

	log.Printf("Getting ip address for container %s", c.Name)

	go func() {
		for retry := timeout; retry > 0; retry -= 1 {
			time.Sleep(time.Second * 1)
			info, err := c.Info()
			if err != nil {
				log.Printf("Error while getting ip address for container %s, %s", c.Name, err.Error())
				ip_chan <- ""
			}
			if ipAddr, ok := info["IP"]; ok {
				log.Printf("Got ip address for container %s, %s", c.Name, ipAddr)
				ip_chan <- ipAddr
				break
			}
		}
	}()

	select {
	case res := <-ip_chan:
		if err := c.Stop(); err != nil {
			return res, err
		}
		return res, nil
	case <-time.After(time.Second * time.Duration(timeout)):
		if err := c.Stop(); err != nil {
			return "", err
		}
		log.Printf("Timeout while getting ip address for container %s", c.Name)
		return "", nil
	}

	return "", nil

}

func (c *Container) Exists() bool {
	_, err := c.Info()
	if err != nil {
		return false
	}
	return true
}

func (c *Container) CGroup(name string) (string, error) {
	out, err := common.RunCommand(path.Join(LXC_BIN, "lxc-cgroup"), []string{"-n", c.Name, name})
	return out, err
}

func (c *Container) CpuUsage() ([]int, error) {
	out, err := c.CGroup("cpuacct.usage_percpu")
	if err != nil {
		return nil, err
	}
	per_cpu_str := strings.Fields(out)
	per_cpu := make([]int, len(per_cpu_str))
	for idx, value := range per_cpu_str {
		int_value, _ := strconv.Atoi(value)
		per_cpu[idx] = int_value
	}
	return per_cpu, err
}

func (c *Container) RamUsage() (int, error) {
	out, err := c.CGroup("memory.usage_in_bytes")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.Trim(out, "\n"))
}

func (c *Container) Create(template string, fssize int, config string) (string, error) {
	out, err := common.RunCommand(path.Join(LXC_BIN, "lxc-create"), []string{"-n", c.Name,
		"-t", template, "-B", "lvm", fmt.Sprintf("--fssize=%vGB", fssize)})
	if err := c.AppendConfig(config); err != nil {
		return out, err
	}
	return out, err
}

func (c *Container) Destroy() (string, error) {
	out, err := common.RunCommand(path.Join(LXC_BIN, "lxc-destroy"), []string{"-n", c.Name})
	return out, err
}

func (c *Container) ResetPassword(user, password string) error {
	log.Printf("Resetting password for container %s", c.Name)

	proc := exec.Command(path.Join(LXC_BIN, "lxc-attach"), "-n", c.Name, "--", "passwd", user)

	stdin, err := proc.StdinPipe()
	if err != nil {
		log.Printf("Error while resetting password for container %s, %s", c.Name, err.Error())
		return err
	}
	defer stdin.Close()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	proc.Stdout = stdout
	proc.Stderr = stderr

	if err := proc.Start(); err != nil {
		log.Printf("Error while resetting password for container %s, %s, %s", c.Name, string(stderr.Bytes()), string(stdout.Bytes()))
		return err
	}

	io.WriteString(stdin, fmt.Sprintf("%s\n%s\n", password, password))
	if err := proc.Wait(); err != nil {

		log.Printf("Error while resetting password for container %s, %s, %s", c.Name, string(stderr.Bytes()), string(stdout.Bytes()))
		return err
	}
	log.Printf("Resetted password for container %s", c.Name)

	return nil

}

func (c Container) String() string {
	return fmt.Sprintf("name: %s", c.Name)
}

func (c *Container) ConfigPath() string {
	return path.Join(LXC_BASE_PATH, c.Name, "config")
}

func (c *Container) ReadConfig() (string, error) {
	path := c.ConfigPath()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), err
}

func (c *Container) AppendConfig(config string) error {
	path := c.ConfigPath()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 644)
	defer file.Close()

	if err != nil {
		return err
	}

	if _, err := file.WriteString(config); err != nil {
		return err
	}

	return nil
}

func (c *Container) ReplaceConfig(config string) error {
	path := c.ConfigPath()
	return ioutil.WriteFile(path, []byte(config), 644)
}

func NewContainer(name string) *Container {
	return &Container{Name: name}
}

func ListContainers() ([]map[string]interface{}, error) {
	out, err := common.RunCommand(path.Join(LXC_BIN, "lxc-ls"), []string{"--fancy", "-F", "name,state"})
	lines := strings.Split(out, "\n")
	res := make([]map[string]interface{}, 0)
	for _, value := range lines[1:] {
		item := make(map[string]interface{})
		values := strings.Fields(value)
		if len(values) > 0 {
			item["Name"] = values[0]
			item["State"] = values[1]
			res = append(res, item)
		}
	}
	return res, err
}
