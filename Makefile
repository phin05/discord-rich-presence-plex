build: build-web build-go
build-web:
	make -C web build
build-go: build-go@windows build-go@linux build-go@darwin
build-go@windows: build-go-windows-amd64
build-go@linux: build-go-linux-amd64 build-go-linux-arm64 build-go-linux-386 build-go-linux-arm
build-go@darwin: build-go-darwin-amd64 build-go-darwin-arm64
build-go-%:
	$(eval export CGO_ENABLED = 0)
	$(eval OSARCH = $(subst -, ,$*))
	$(eval export GOOS = $(word 1,$(OSARCH)))
	$(eval export GOARCH = $(word 2,$(OSARCH)))
	@echo Building $*
	go build -ldflags "-s -w" -trimpath -o out/$(GOOS)-$(GOARCH)/
	cp LICENSE COPYRIGHT NOTICE README.md out/$(GOOS)-$(GOARCH)/

clean: clean-web clean-go
clean-web:
	make -C web clean
clean-go:
	rm -rf out

dev:
	nodemon --signal SIGKILL --ext go --exec "(go build -o out\dev.exe && out\dev.exe) || exit 1"

test:
	go test -v ./...
test-coverage:
	go test -coverprofile=coverage.txt -v ./...
	go tool cover -html=coverage.txt
	rm -f coverage.txt

vet:
	-go vet ./...
	-errcheck -asserts -blank ./...
	-nilaway ./...
	-gosec -quiet ./...
	-staticcheck ./...
