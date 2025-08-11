TARGET = fzz
BIN_DIR = bin

# OS conditions
ifeq ($(OS),Windows_NT)
	TARGET_NAME := $(TARGET).exe
	RM := rmdir /s /q
else
	TARGET_NAME := $(TARGET)
	RM := rm -rf
endif

.PHONY: build clean

# builds
build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(TARGET_NAME) ./fzz.go
	@echo "Built in $(BIN_DIR)/$(TARGET_NAME)"

# cleans
clean:
	@$(RM) $(BIN_DIR)
	@echo "Clean now"
