BUILD_FLAGS = -ldflags="-s -w"

build:
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o bootstrap
	@zip function.zip bootstrap
fmt:
	@go fmt ./...
