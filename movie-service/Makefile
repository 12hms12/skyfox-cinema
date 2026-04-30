# Movie Service Makefile

appVersion=1.0-build

ARTIFACTORY_DOCKER_REGISTRY = #private registry
IMAGE_NAME = movie-service

build:
	GOPROXY=https://proxy.golang.org,direct go build -o out/server ./main.go

run:
	GOPROXY=https://proxy.golang.org,direct go run main.go

deploy: build run-binary

run-binary:
	./out/server

deps:
	GOPROXY=https://proxy.golang.org,direct go mod tidy
	GOPROXY=https://proxy.golang.org,direct go mod download

clean:
	rm -rf out/

static-checks: fmt lint

fmt:
	go fmt ./...
	go vet ./...

lint:
	#please install golangci-lint if not installed, brew install golangci-lint
	golangci-lint run -v

test:
	go clean -testcache
	mkdir -p out/
	go test ./... -coverprofile=out/coverage.out
	go tool cover -html=out/coverage.out -o out/coverage.html

docker/login:
	@docker login -u $(ARTIFACTORY_HA_USERNAME) -p $(ARTIFACTORY_HA_PASSWORD) $(ARTIFACTORY_DOCKER_REGISTRY)

docker/build: docker/login
	docker build -f Dockerfile.go -t ${IMAGE_NAME} .

docker/run: docker/build
	docker run -p 4567:4567 ${IMAGE_NAME}

docker/publish: docker/login
	docker tag ${IMAGE_NAME} $(ARTIFACTORY_DOCKER_REGISTRY)/${IMAGE_NAME}:$(appVersion)
	docker push $(ARTIFACTORY_DOCKER_REGISTRY)/${IMAGE_NAME}:$(appVersion)

.PHONY: build run deploy run-binary deps clean fmt lint test docker/login docker/build docker/run docker/publish
