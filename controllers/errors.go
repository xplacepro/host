package controllers

import (
	"errors"
	"fmt"
)

func NotStopped(hostname string) error {
	return errors.New(fmt.Sprintf("%s is not in STOPPED state", hostname))
}

var ValidationError = errors.New("validation error")
var AlreadyExistsError = errors.New("already exists")
