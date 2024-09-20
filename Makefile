.PHONY: all build run clean

BINARY_NAME=bot

all: build

build:
	@echo "Building the project..."
	go build -o $(BINARY_NAME) src/main.go src/cron.go src/csv.go src/issues.go src/projects.go src/slack.go src/utils.go src/webhook.go

run: build
	@echo "Running the project..."
	./$(BINARY_NAME)

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)