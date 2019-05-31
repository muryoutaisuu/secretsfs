GO15VENDOREXPERIMENT=1
GO111MODULE="off"

all: clean build

clean:
	if [ -f secretsfs ] ; then rm secretsfs ; fi
	if [ -d dist ] ; then rm -rf dist ; fi

build:
	go get -v -u ./...
	goreleaser
