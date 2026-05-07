.PHONY: build build-all build-linux build-windows build-darwin clean test

build:
	go build -o bin/acr-uploader main.go

build-all: build-linux build-windows build-darwin

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/acr-uploader-linux-amd64 main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/acr-uploader-windows-amd64.exe main.go

build-darwin:
	GOOS=darwin GOARCH=arm64 go build -o bin/acr-uploader-darwin-arm64 main.go

clean:
	rm -rf bin/

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
