GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

GO_OPT= -mod vendor -ldflags "-X main.Branch=$(GIT_BRANCH) -X main.Revision=$(GIT_REVISION) -X main.Version=$(VERSION)"

.PHONY: prototype
prototype:
	GO111MODULE=on CGO_ENABLED=0 go build $(GO_OPT) -o ./bin/$(GOOS)/prototype-$(GOARCH) $(BUILD_INFO) ./cmd/prototype

.PHONY: protoctl
protoctl:
	GO111MODULE=on CGO_ENABLED=0 go build $(GO_OPT) -o ./bin/$(GOOS)/protoctl-$(GOARCH) $(BUILD_INFO) ./cmd/protoctl

-PHONY: docker
docker:
	docker build -f cmd/prototype/Dockerfile -t ghcr.io/dgzlopes/prototype:latest .

-PHONY: docker-protoctl
docker-protoctl:
	docker build -f cmd/protoctl/Dockerfile -t ghcr.io/dgzlopes/protoctl:latest .

.PHONY: integration
integration:
	pip3 install -r integration/requirements.txt
	python3 integration/e2e.py