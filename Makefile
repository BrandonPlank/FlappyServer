all: build

build:
	@echo "Building server for GNU/Linux"
	@go build -v
	@echo "Making the binary executable"
	@chmod +x FlappyServer
run:
	go run .

