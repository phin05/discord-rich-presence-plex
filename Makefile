APP_VERSION = 3.0.0

MAKEFLAGS += --no-print-directory

export GOEXPERIMENT := jsonv2

build: build-web build-server
build-web:
	APP_VERSION=$(APP_VERSION) make -C web build
	mkdir web/dist/static
	cp NOTICE.md README.md LICENSE COPYRIGHT web/dist/static/
build-server: build-server@windows build-server@linux build-server@darwin
build-server@windows: build-server-windows-amd64
build-server@linux: build-server-linux-amd64 build-server-linux-arm64 build-server-linux-386 build-server-linux-arm
build-server@darwin: build-server-darwin-amd64 build-server-darwin-arm64
build-server-%:
	$(eval export CGO_ENABLED := 0)
	$(eval OSARCH := $(subst -, ,$*))
	$(eval export GOOS := $(word 1,$(OSARCH)))
	$(eval export GOARCH := $(word 2,$(OSARCH)))
	go build -ldflags "-buildid= -s -w$(if $(filter windows,$(GOOS)), -H windowsgui,) -X main.version=$(APP_VERSION)" -trimpath -o out/$(GOOS)-$(GOARCH)/$(if $(filter windows,$(GOOS)),DRPP.exe,drpp) ./server/main

clean: clean-web clean-server
clean-web:
	make -C web clean
clean-server:
	rm -rf out

install-deps: install-deps-web install-deps-server
install-deps-web:
	make -C web install-deps
install-deps-server:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@b16c91e2c891bd4a1234508919f1a66682a2284b

upgrade-deps: upgrade-deps-web upgrade-deps-server
upgrade-deps-web:
	make -C web upgrade-deps
upgrade-deps-server:
	go get -u ./...
	go mod tidy

dev:
	nodemon --signal SIGKILL --ext go --exec "(go build -tags dev -o out/dev.exe ./server/main && out\dev.exe) || exit 1"

test:
	go test -count=1 -v ./...

lint:
	@err=0; \
	make lint-web || err=1; \
	make lint-server || err=1; \
	exit $$err
lint-web:
	make -C web lint
lint-server:
	golangci-lint run --allow-parallel-runners
