module kayladev.me/FlappyServer/models

go 1.17

require (
	github.com/google/uuid v1.3.0
	github.com/jinzhu/gorm v1.9.16
	golang.org/x/crypto v0.0.0-20211209193657-4570a0811e8b
	kayladev.me/FlappyServer/database v0.0.0
)

replace kayladev.me/FlappyServer/database => ./../database
