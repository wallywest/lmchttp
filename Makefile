BUILD_PATH=github.com/wallywest/lmchttp/cmd/lmchttp

build:
	@mkdir -p bin/
	@govendor build -o bin/lmchttp $(BUILD_PATH)
