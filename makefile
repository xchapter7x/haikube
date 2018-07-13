# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=hk
BINARY_DIR=build
BINARY_WIN=$(BINARY_NAME).exe
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_DARWIN=$(BINARY_NAME)_osx

all: test build
build: build-darwin build-win build-linux 
test: unit integration	
unit: 
	$(GOTEST) ./pkg/... -v
integration: 
	$(GOTEST) ./test/integration/... -v
e2e: 
	$(GOTEST) ./test/e2e/... -v
clean: 
	$(GOCLEAN)
	rm -fr $(BINARY_DIR)
dep:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure
build-darwin: 
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -v -o $(BINARY_DIR)/$(BINARY_DARWIN) ./cmd/haikube
build-win:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -v -o $(BINARY_DIR)/$(BINARY_WIN) ./cmd/haikube
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -v -o $(BINARY_DIR)/$(BINARY_UNIX) ./cmd/haikube

.PHONY: all test clean build
