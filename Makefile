.PHONY: build, build-local, format, docs, install, test, testacc, update-sdk
LOCAL_TEST_VERSION = 99.0.0
OS_ARCH = darwin_amd64

build:
	go build

build-local:
	go build -o terraform-provider-amplience_${LOCAL_TEST_VERSION}
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/labd/amplience/${LOCAL_TEST_VERSION}/${OS_ARCH}
	cp terraform-provider-amplience_${LOCAL_TEST_VERSION} ~/.terraform.d/plugins/registry.terraform.io/labd/amplience/${LOCAL_TEST_VERSION}/${OS_ARCH}/terraform-provider-amplience_v${LOCAL_TEST_VERSION}

format:
	go fmt ./...

docs:
	go generate ./...

install:
	go install .

test:
	go test -v ./...

testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

update-sdk:
	GO111MODULE=on go get github.com/labd/go-apollostudio-sdk
	GO111MODULE=on go mod tidy