all: build

build:
	go build
	chmod +x FlappyServer
run:
	go run .

