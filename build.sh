#!/bin/bash
# set -x

BIN_NAME=microtest:go
DOCKER_NAME=$BIN_NAME

echo "Check go fmt"
cmd="gofmt -l . | grep -v 'vendor/'"
if [[ $(eval $cmd | wc -l) -ne 0 ]]; then
	echo 'Need gofmt for:'
	eval $cmd
	exit 1
fi

echo "Run go vet"
go vet $(go list ./... | grep -v '/vendor/')
if [[ $? -ne 0 ]]; then
	echo 'Error on vet code'
	exit 1
fi

echo "Check golint"
cmd="golint ./... | grep -v 'vendor/' | grep -v 'should have comment' | grep -v 'underscore in package name'"
if [[ $(eval $cmd | wc -l) -ne 0 ]]; then
	echo 'Golint have some problems:'
	eval $cmd
	exit 1
fi

echo "Run tests"
go test ./... >/dev/null
if [[ $? -ne 0 ]]; then
	echo 'Error on run tests:'
	go test ./...
	exit 1
fi

go install

echo "Build app"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

if [[ $? -ne 0 ]]; then
	echo 'Error on build'
	exit 1
fi

echo "Create docker image"
docker build -t $DOCKER_NAME .
echo "Docker image:" $DOCKER_NAME

rm app
