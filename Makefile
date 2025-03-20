APP_NAME = creek
BUILD_DIR = target

GO_OS ?= $(shell go env GOOS)
GO_ARCH ?= $(shell go env GOARCH)

# Detect OS (Windows uses del instead of rm)
RM = rm -rf
ifeq ($(OS), Windows_NT)
    RM = del /S /Q
endif

.PHONY: all clean build test

all: clean build

ifeq ($(OS), Windows_NT)
    MKDIR = if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
else
    MKDIR = mkdir -p $(BUILD_DIR)
endif

# Check if Windows (needs .exe extension)
ifeq ($(GO_OS), windows)
    OUTPUT_FILE = $(BUILD_DIR)/$(APP_NAME)-$(GO_OS)-$(GO_ARCH).exe
else
    OUTPUT_FILE = $(BUILD_DIR)/$(APP_NAME)-$(GO_OS)-$(GO_ARCH)
endif

# Build the project for the current OS
build:
	@echo "Building $(APP_NAME) for $(GO_OS)/$(GO_ARCH)..."
	@$(MKDIR)
	go mod tidy
	go build -o $(OUTPUT_FILE) cmd/server/main.go
	@echo "Build completed: $(BUILD_DIR)/$(APP_NAME)-$(GO_OS)-$(GO_ARCH)"

# Build for all platforms
build-all:
	@echo "Building for all major platforms..."
	@$(MKDIR)
	GOOS=linux  go build -o $(BUILD_DIR)/$(APP_NAME)-linux cmd/server/main.go
	GOOS=darwin  go build -o $(BUILD_DIR)/$(APP_NAME)-mac cmd/server/main.go
	GOOS=windows  go build -o $(BUILD_DIR)/$(APP_NAME)-windows.exe cmd/server/main.go
	@echo "Cross-platform builds completed in $(BUILD_DIR)/"

# Run tests
test:
	go test ./test

# Clean build directory
clean:
	@echo "Cleaning up..."
	-$(RM) $(BUILD_DIR)  # The '-' prevents errors if dir doesn't exist
	@echo "Cleaned."
