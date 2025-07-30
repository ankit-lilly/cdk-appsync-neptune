BUILD_FLAGS = -ldflags="-s -w"

.PHONY: build
build:
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o bootstrap
	@zip function.zip bootstrap

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: clean
clean:
	@rm -f bootstrap function.zip
