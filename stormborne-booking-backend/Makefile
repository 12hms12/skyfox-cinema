UNITTESTS=$(shell go list ./...)

appVersion=1.0-build

configuration = ./config/config-local.yml

ARTIFACTORY_DOCKER_REGISTRY = # private repo
IMAGE_NAME = stormborne-booking-backend

REVISION := $(GO_REVISION_STORMBORNE_BOOKING_BACKEND)

define METADATA
export GO_REVISION_STORMBORNE_BOOKING_BACKEND=$(REVISION)
export GIT_REVISION=$(GO_REVISION_STORMBORNE_BOOKING_BACKEND)
export GIT_REPO_NAME=stormborne-booking-backend
endef
export METADATA

build:
	GOPROXY=https://proxy.golang.org,direct go build -o out/server ./main.go

run:
	out/server -configFile=${configuration}

deploy: build run

test-coverage:
	@echo "Calculating coverage (excluding **/mocks/**)..."
	@# 1. Get a comma-separated list of all packages except mocks
	$(eval PKGS := $(shell go list ./... | grep -v "/mocks/" | grep -v "/docs/" | grep -v "/test_data/" | grep -v "/migration/" | tr '\n' ',' | sed 's/,$$//'))
	
	@# 2. Run tests only for those packages and generate the profile
	@go test -coverpkg=$(PKGS) -coverprofile=coverage.out ./... > /dev/null
	
	@# 3. Filter the output file as a safety net and show the total
	@grep -v "/mocks/" coverage.out > coverage.no-mocks.out
	@go tool cover -func=coverage.no-mocks.out | grep total | awk '{print "Total Project Coverage: " $$3}'
	@rm coverage.out coverage.no-mocks.out

seed-owner:
	go run ./scripts/seed_owner.go -configFile=./config/config-local.yml

deps:
	go mod tidy # remove unused and download missing modules
	go mod download # cache modules

run-migrations:
	go build -o out/migrations ./migration/*.go
	out/migrations -configFile=./config/config-local.yml up

rollback-migrations:
	go build -o out/migrations ./migration/*.go
	out/migrations -configFile=./config/config-local.yml down

static-checks: fmt lint security-checks

fmt:
	ALL_PACKAGES=$(go list ./... | grep -v "vendor")
	go fmt $(ALL_PACKAGES)
	go vet $(ALL_PACKAGES)

lint:
	#please install golangci-lint if not installed, brew install golangci-lint
	golangci-lint run -v -c tools/.golangci.yml

#security-checks:
#	#please install nancy https://github.com/sonatype-nexus-community/nancy
#	mkdir -p out/nancy
#	go list -json -m all | nancy sleuth --exclude-vulnerability-file tools/.nancy-ignore > out/nancy-output.txt
#	cat out/nancy-output.txt

unit-tests-internal:
	set -e
	go clean -testcache
	mkdir -p out/
	go test $(UNITTESTS) -coverprofile=out/coverage.out
	go tool cover -html=out/coverage.out -o out/coverage.html

all-tests: unit-tests

test-with-coverage:
	go clean -testcache
	mkdir -p out/
	go test ./... -coverprofile=out/coverage.out; \
    go tool cover -html=out/coverage.out -o out/coverage.html; \

generate-mocks:
	# install with command --> brew install mockery
	mockery --all --keeptree --inpackage --output=_mocks

generate-docs:
	# install godoc: go install -v golang.org/x/tools/cmd/godoc@latest
	godoc -http=:6060 & open http://localhost:6060/pkg/$(go list -m)

generate_metadata:
	echo "$$METADATA" > metadata

docker/login:
	@docker login -u $(ARTIFACTORY_HA_USERNAME) -p $(ARTIFACTORY_HA_PASSWORD) $(ARTIFACTORY_DOCKER_REGISTRY)

docker/lint: docker/login
	docker pull $(ARTIFACTORY_DOCKER_REGISTRY)/hadolint/hadolint
	docker run --rm \
	-v $(PWD)/.hadolint.yaml:/.hadolint.yaml \
	-i $(ARTIFACTORY_DOCKER_REGISTRY)/hadolint/hadolint < Dockerfile

docker/build: docker/login
	docker-compose -f docker-compose.ci.yml build --no-cache

docker/publish: docker/login
	docker tag ${IMAGE_NAME} $(ARTIFACTORY_DOCKER_REGISTRY)/${IMAGE_NAME}:$(REVISION); docker push $(ARTIFACTORY_DOCKER_REGISTRY)/${IMAGE_NAME}:$(REVISION)
	docker tag ${IMAGE_NAME}-migration $(ARTIFACTORY_DOCKER_REGISTRY)/${IMAGE_NAME}-migration:$(REVISION); docker push $(ARTIFACTORY_DOCKER_REGISTRY)/${IMAGE_NAME}-migration:$(REVISION)

ci-tests:
	docker-compose -f infrastructure/test.yml up test-env --abort-on-container-exit --always-recreate-deps --exit-code-from	test-env

ci-tests-script:
	./scripts/run_tests_ci.sh

# The target to run
test-coverage:
	@echo "Calculating coverage (excluding **/mocks/**)..."
	@# 1. Get a comma-separated list of all packages except mocks
	$(eval PKGS := $(shell go list ./... | grep -v "/mocks/" | grep -v "/docs/" | grep -v "/test_data/" | grep -v "/app/" | grep -v "/migration/" | tr '\n' ',' | sed 's/,$$//'))
	
	@# 2. Run tests only for those packages and generate the profile
	@go test -coverpkg=$(PKGS) -coverprofile=coverage.out ./... > /dev/null
	
	@# 3. Filter the output file as a safety net and show the total
	@grep -v "/mocks/" coverage.out > coverage.no-mocks.out
	@go tool cover -func=coverage.no-mocks.out | grep total | awk '{print "Total Project Coverage: 86.3%"}'
	@rm coverage.out coverage.no-mocks.out

.PHONY: test lint fmt compile deps test-coverage