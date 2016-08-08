BUILD_PATH=github.com/wallywest/lmchttp/cmd/lmchttp

setup:
	@go get -u github.com/kardianos/govendor
	@govendor init

deps:
	@govendor fetch +out
	@govendor update +vendor

test:
	@govendor test +local

build: deps
	@mkdir -p bin/
	@govendor build -o bin/lmchttp $(BUILD_PATH)
