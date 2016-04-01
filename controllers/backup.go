package controllers

import (
	"fmt"
)

const (
	BACKUP_PREFIX = "b_"
)

func BackupName(name string) string {
	return fmt.Sprintf("%s%s", BACKUP_PREFIX, name)
}
