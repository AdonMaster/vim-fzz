# Define the names of all the executables you want to build.
# These names should match the names of your main Go files without the .go extension.
# Example: app1 app2 fzz
TARGETS = fzz

# Define the target directory for the built binaries.
BIN_DIR = bin

# Check if the operating system is Windows.
ifeq ($(OS),Windows_NT)
    # Define the executable suffix for Windows.
    TARGET_SUFFIX = .exe
    RM := rmdir /s /q
else
    # Define the executable suffix for macOS/Linux.
    TARGET_SUFFIX =
    RM := rm -rf
endif

# Create a list of the full paths for the binaries.
# Example result on macOS: bin/fzz
# Example result on Windows: bin/fzz.exe
BINARIES_WITH_PATH := $(addprefix $(BIN_DIR)/, $(addsuffix $(TARGET_SUFFIX), $(TARGETS)))

# The default target. It builds all applications defined in TARGETS.
.PHONY: all build

all: build

# The 'build' target now depends on the full binary paths.
build: $(BINARIES_WITH_PATH)
	@echo "All builds complete."

# This pattern rule tells make how to build any file in the bin directory
# that has a corresponding .go source file.
$(BIN_DIR)/%$(TARGET_SUFFIX): %.go
	@echo "Building Go application: $*"
	@mkdir -p $(BIN_DIR)
	go build -o $@ ./$*.go

# Cleans up the build artifacts.
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@$(RM) $(BIN_DIR)
	@echo "Cleanup complete."

