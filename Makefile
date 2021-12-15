all: build

build:
	echo "Building server for GNU/Linux"
	go build -v
	chmod +x FlappyServer
run:
	go run .

