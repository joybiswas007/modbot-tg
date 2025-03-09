# Go parameters
GOCMD=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=modbot

# Run the bot
run:
	@echo "Running the bot..."
	$(GORUN) ./cmd/bot/

# Build the bot
build:
	@echo "Building the bot..."
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/bot/

# Test the application
test:
	@echo "Testing..."
	$(GOTEST) -v ./...

# Clean up build files
clean:
		@echo "Cleaning..."
		$(GOCLEAN)
		rm -rf $(BINARY_NAME)
