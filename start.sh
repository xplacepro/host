#!/bin/sh
# echo "arguments: $*" > /tmp/test
# echo "environment:" >> /tmp/test
# env | grep LXC >> /tmp/test

exec $GOPATH/bin/host --config= --notify=RUNNING --hostname=$1