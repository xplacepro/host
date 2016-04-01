package lvm2

import (
	"strings"
)

func NotFound(err error) bool {
	return strings.Contains(err.Error(), "not found")
}

func AlreadyExists(err error) bool {
	return strings.Contains(err.Error(), "already exists in volume group")
}

func AlreadyRemoved(err error) bool {
	return strings.Contains(err.Error(), "One or more specified logical volume(s) not found")
}
