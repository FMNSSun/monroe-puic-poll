#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

CONTAINER=${DIR##*/}
DOCKERFILE=${CONTAINER}.docker

echo "Compiling puic-poll.go"

env GOOS=linux GOARCH=amd64 go build ./puic-poll.go
chmod +x ./puic-poll
cp -f ./puic-poll ./files/puic-poll

docker pull monroe/base
docker build --rm=true -f ${DOCKERFILE} -t ${CONTAINER} . && echo "Finished building ${CONTAINER}"
