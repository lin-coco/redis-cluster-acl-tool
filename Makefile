# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

# Project name
BINARY_NAME = acltool

# Main build target
all: build

# Build the project
build:
	go mod tidy
	$(GOBUILD) -o $(BINARY_NAME) -v

# Build for Linux (cross-compiling)
build-amd64:
	go mod tidy
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME) -v

# Test the project
test:
	$(GOTEST) -v ./...

# Get project dependencies
get:
	$(GOGET) ./...

# Clean the project
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run the project
run:
	$(GOBUILD) -o $(BINARY_NAME) -v
	./$(BINARY_NAME)

# Default target
.DEFAULT_GOAL := all
