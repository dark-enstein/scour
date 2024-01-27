DOCKER_USERNAME ?= brownbarg
APPLICATION_NAME ?= scour
APPLICATION_DIR := ./build
APPLICATION_FILEPATH := ./build/scour
REFVER := $(shell cat .git/HEAD | cut -d "/" -f 3)
SEMVER ?= $(shell cat .git/HEAD | cut -d "/" -f 3)
HELMPACKAGE := "bumper"

.PHONY: build test clean run docker-run docker-build docker-push docker-tag helm-create helm-package helm-test kdeploy

test:
	go test ./... -v
	sh script/mshift_test.sh

build:
	rm -rf build
	@echo $(REFVER)
	@CGO_ENABLED=0 go build -v --o $(APPLICATION_FILEPATH) main.go

fmt:
	@go fmt

clean:
	@rm -rf ./build

run: build
	@echo "Running with args: $(ARGS)"
	@./build/scour $(ARGS)

docker-build:
	docker build . -t $(APPLICATION_NAME):$(REFVER)

docker-tag: docker-build
	@docker tag $(APPLICATION_NAME):$(REFVER) brownbarg/$(APPLICATION_NAME):$(SEMVER)

docker-run-it: docker-build docker-tag
	docker run -it $(APPLICATION_NAME):$(REFVER) /bin/sh

docker-run:
	docker run $(APPLICATION_NAME):$(REFVER) scour -X GET "https://httpbin.org/get" -H "accept: application/json" -v

docker-push: docker-tag
	@echo $$DOCKER_PASSWORD | docker login -u $(DOCKER_USERNAME) --password-stdin
	@docker push brownbarg/$(APPLICATION_NAME):$(SEMVER)