module brandonplank.org/FlappyServer/routes

go 1.17

require (
	brandonplank.org/FlappyServer/database v0.0.0
	brandonplank.org/FlappyServer/global v0.0.0
	brandonplank.org/FlappyServer/models v0.0.0
	github.com/gofiber/fiber/v2 v2.23.0
	github.com/google/uuid v1.3.0
	github.com/jinzhu/gorm v1.9.16
	golang.org/x/crypto v0.0.0-20211209193657-4570a0811e8b
)

require (
	github.com/andybalholm/brotli v1.0.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/klauspost/compress v1.13.4 // indirect
	github.com/mattn/go-sqlite3 v1.14.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.31.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
)

replace (
	brandonplank.org/FlappyServer/database => ./../Database
	brandonplank.org/FlappyServer/models => ./../Models
	brandonplank.org/FlappyServer/global => ./../Global
)
