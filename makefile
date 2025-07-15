APP_NAME := ImageUtil
SRC_DIR := src
BUILD_DIR := builds

.PHONY: all build clean linux windows darwin local

all: clean build

build: linux windows darwin

linux:
	cd $(SRC_DIR) && GOOS=linux GOARCH=amd64 go build -o ../$(BUILD_DIR)/$(APP_NAME)-linux-amd64

windows:
	cd $(SRC_DIR) && GOOS=windows GOARCH=amd64 go build -o ../$(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe

local:
	cd $(SRC_DIR) && go build -o ../$(BUILD_DIR)/$(APP_NAME)

clean:
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)

