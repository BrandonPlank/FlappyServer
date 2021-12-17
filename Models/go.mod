module brandonplank.org/FlappyServer/models

go 1.17

require (
	github.com/google/uuid v1.3.0
	github.com/jinzhu/gorm v1.9.16
	golang.org/x/crypto v0.0.0-20211209193657-4570a0811e8b
	brandonplank.org/FlappyServer/database v0.0.0
)

replace brandonplank.org/FlappyServer/database => ./../Database
