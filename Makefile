APP_VERSION := 3.3.1

export GOEXPERIMENT := jsonv2

build: build-web build-server
build-web:
	@APP_VERSION=$(APP_VERSION) "$(MAKE)" -C web build
	mkdir web/dist/static
	cp NOTICE.md README.md LICENSE COPYRIGHT web/dist/static/
build-server: build-server@windows build-server@linux build-server@darwin
build-server@windows: build-server-windows-amd64
build-server@linux: build-server-linux-amd64 build-server-linux-arm64 build-server-linux-386 build-server-linux-arm
build-server@darwin: build-server-darwin-amd64 build-server-darwin-arm64
build-server-%:
	$(eval OSARCH := $(subst -, ,$*))
	$(eval GOOS := $(word 1,$(OSARCH)))
	$(eval GOARCH := $(word 2,$(OSARCH)))
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-buildid= -s -w -X main.version=$(APP_VERSION)$(if $(filter windows,$(GOOS)), -H windowsgui,)" -trimpath -o out/$(GOOS)-$(GOARCH)/drpp$(if $(filter windows,$(GOOS)),.exe,) ./server/main

clean: clean-web clean-server
clean-web:
	@"$(MAKE)" -C web clean
clean-server:
	rm -rf out

install-deps: install-deps-web install-deps-server
install-deps-web:
	@"$(MAKE)" -C web install-deps
install-deps-server:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@b16c91e2c891bd4a1234508919f1a66682a2284b

upgrade-deps: upgrade-deps-web upgrade-deps-server
upgrade-deps-web:
	@"$(MAKE)" -C web upgrade-deps
upgrade-deps-server:
	go get -u ./...
	go mod tidy

dev:
	nodemon --signal SIGKILL --ext go --exec "go build -tags dev -o out/dev.exe ./server/main && out\dev.exe"

lint:
	@"$(MAKE)" -j lint-web lint-server
lint-web:
	@"$(MAKE)" -C web lint
lint-server:
	golangci-lint run --allow-parallel-runners ./server/... ./web

test:
	go test -count=1 -v ./...
