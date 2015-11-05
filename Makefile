PROG_NAME := "bitindex"
GIT_VERSION := $(shell git log -1 --pretty=format:"%h (%ci)" .)

build:
	go build -ldflags "-X \"main.buildVersion=$(GIT_VERSION)\"" \
		-o $(GOPATH)/bin/$(PROG_NAME)

dist:
	mkdir -p dist

	gox -ldflags "-X \"main.buildVersion=$(GIT_VERSION)\"" \
		-os "darwin linux windows" \
		-arch "amd64" \
		-output="./dist/$(PROG_NAME)-{{.OS}}-{{.Arch}}" \

install:
	go get github.com/mitchellh/gox
	go get golang.org/x/tools/cmd/cover
	go get github.com/spf13/viper
	go get github.com/spf13/cobra
	go get github.com/blang/semver

test:
	go test -v -cover -bench . -benchmem .

.PHONY: dist test
